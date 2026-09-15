package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
)

/*
	// Приведение интерфейса к конкретному типу Connector для извлечения настроек
	localConn, ok := r.opts.Connector.(*Connector)
	if !ok {
		return nil, fmt.Errorf("invalid connector type, expected *kafka.Connector")
	}

	newReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        localConn.GetBrokers(),
		Topic:          r.opts.TargetName,
		GroupID:        r.opts.GroupID,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

*/

type Receiver struct {
	opts   *ReceiverOptions
	reader KafkaReceiverLink
	logger logger.Logger
	mu     sync.RWMutex
	initMu sync.Mutex
}

var _ pkgamqp.Receiver[any] = (*Receiver)(nil)

func NewReceiver(opts ...ReceiverOption) (*Receiver, error) {
	clientOpts := NewReceiverOptions()
	for _, opt := range opts {
		opt(clientOpts)
	}
	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewReceiver", "kafka receiver options validation failed", err)
	}
	return &Receiver{
		opts:   clientOpts,
		logger: clientOpts.Logger.GetLogger("kafka-receiver"),
	}, nil
}

func (r *Receiver) Receive(ctx context.Context, _ any) (pkgamqp.Message, error) {
	kafkaReader, err := r.getReceiver(ctx)
	if err != nil {
		return nil, errs.NewTlCommonError("Receive", "kafka receiver failed to get reader", err)
	}

	kafkaMsg, err := kafkaReader.FetchMessage(ctx)
	if err != nil {
		r.handleReceiverFailure(err)
		return nil, errs.NewTlCommonError("Receive", "kafka incoming packet error", err)
	}

	var finalPayload []byte
	if len(kafkaMsg.Value) > 0 {
		finalPayload = make([]byte, len(kafkaMsg.Value))
		copy(finalPayload, kafkaMsg.Value)
	}

	resMsg := &Message{
		Payload:    finalPayload,
		Props:      make(map[string]any),
		TargetName: r.opts.TargetName,
	}
	if len(kafkaMsg.Key) > 0 {
		resMsg.Props["kafka_message_key"] = string(kafkaMsg.Key)
	}

	// ИСПРАВЛЕНИЕ: Гарантируем уникальную аллокацию структуры в куче для защиты от перезаписи в цикле
	allocatedMsg := new(kafka.Message)
	*allocatedMsg = kafkaMsg
	resMsg.Props[sysMsgKey] = allocatedMsg

	return resMsg, nil
}

func (r *Receiver) Accept(ctx context.Context, msg pkgamqp.Message) error {
	kafkaMsgPtr, err := ExtractOriginalKafkaMessage(msg)
	if err != nil {
		return err
	}

	kafkaReader, err := r.getReceiver(ctx)
	if err != nil {
		return err
	}

	if err = kafkaReader.CommitMessages(ctx, *kafkaMsgPtr); err != nil {
		return errs.NewTlCommonError("Accept", "kafka failed to commit offset", err)
	}
	return nil
}

func (r *Receiver) Reject(ctx context.Context, msg pkgamqp.Message, err error) error {
	r.logger.Errorf("Message processing rejected: %v. Moving offset forward.", err)
	return r.Accept(ctx, msg)
}

func (r *Receiver) Release(ctx context.Context, msg pkgamqp.Message) error {
	r.logger.Warnf("Kafka release called: message will be re-read upon partition rebalance.")
	return nil
}

func (r *Receiver) Close(ctx context.Context) error {
	r.mu.Lock()
	readerToClose := r.reader
	r.reader = nil
	r.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, r.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)
	go func() {
		var closeErrs []error
		if !utils.IsNil(readerToClose) {
			if err := readerToClose.Close(); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}
		done <- errors.Join(closeErrs...)
	}()

	select {
	case err := <-done:
		if err != nil {
			return errs.NewTlCommonError("Close", "kafka receiver close fails", err)
		}
		return nil
	case <-closeCtx.Done():
		return errs.NewTlCommonError("Close", "kafka receiver close timeout", closeCtx.Err())
	}
}

func (r *Receiver) GetTargetName() string { return r.opts.TargetName }

//goland:noinspection DuplicatedCode
func (r *Receiver) getReceiver(ctx context.Context) (KafkaReceiverLink, error) {
	r.mu.RLock()
	if !utils.IsNil(r.reader) {
		defer r.mu.RUnlock()
		return r.reader, nil
	}
	r.mu.RUnlock()

	r.initMu.Lock()
	defer r.initMu.Unlock()

	r.mu.RLock()
	if !utils.IsNil(r.reader) {
		defer r.mu.RUnlock()
		return r.reader, nil
	}
	r.mu.RUnlock()

	_, err := r.opts.Connector.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	// Приведение интерфейса к конкретному типу Connector для извлечения настроек
	localConn, ok := r.opts.Connector.(*Connector)
	if !ok {
		return nil, fmt.Errorf("invalid connector type, expected *kafka.Connector")
	}

	newReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        localConn.GetBrokers(),
		Topic:          r.opts.TargetName,
		GroupID:        r.opts.GroupID,
		MinBytes:       r.opts.MinBytes,
		MaxBytes:       r.opts.MaxBytes,
		MaxWait:        r.opts.MaxWait,
		CommitInterval: 0,
	})

	r.mu.Lock()
	r.reader = newReader
	r.mu.Unlock()

	return newReader, nil
}

func (r *Receiver) handleReceiverFailure(err error) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return
	}
	r.logger.Warnf("Kafka packet reading failure detected: %v. Notifying connector...", err)
	r.opts.Connector.Invalidate(err)

	r.mu.Lock()
	oldReader := r.reader
	r.reader = nil
	r.mu.Unlock()

	// ОПТИМИЗАЦИЯ: Мягко тушим упавший ридер в фоне, чтобы предотвратить Goroutine Leak
	if !utils.IsNil(oldReader) {
		go func(rd KafkaReceiverLink) {
			_ = rd.Close()
		}(oldReader)
	}
}
