package kafka

import (
	"context"
	"errors"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
)

type Receiver struct {
	opts        *ReceiverOptions
	reader      KafkaReceiverLink
	logger      logger.Logger
	kafkaLogger *logger.KafkaLogger
	mu          sync.RWMutex
	initMu      sync.Mutex
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

	log := clientOpts.Logger.GetLogger("kafka-receiver")

	return &Receiver{
		opts:        clientOpts,
		logger:      log,
		kafkaLogger: logger.NewKafkaLogger(log),
	}, nil
}

func (r *Receiver) Receive(ctx context.Context, _ any) (pkgamqp.Message, error) {
	receiverLink, err := r.getReceiver(ctx)
	if err != nil {
		return nil, errs.NewTlCommonError("Receive", "kafka receiver failed to get reader", err)
	}

	kafkaMsg, err := receiverLink.FetchMessage(ctx)
	if err != nil {
		// Сетевые ошибки FetchMessage обрабатывать Invalidate-ом больше не нужно:
		// kafka.Reader восстанавливает коннекты сам в бэкграунде.
		return nil, errs.NewTlCommonError("Receive", "kafka incoming packet error", err)
	}

	var finalPayload []byte
	if len(kafkaMsg.Value) > 0 {
		finalPayload = make([]byte, len(kafkaMsg.Value))
		copy(finalPayload, kafkaMsg.Value)
	}

	// Использован локальный конструктор NewMessage из нашего пакета (из message.go),
	// вместо ручной сборки. Это делает логику маппинга заголовков и сохранения оригинала единой.
	resMsg := NewMessage(kafkaMsg)
	resMsg.TargetName = r.opts.TargetName

	if len(kafkaMsg.Key) > 0 {
		resMsg.Props["kafka_message_key"] = string(kafkaMsg.Key)
	}

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

// Double-Checked Locking теперь работает напрямую с пулом брокеров из опций
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

	// Инициализируем высокоуровневый пуллер напрямую из списка адресов
	newReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        r.opts.Brokers,
		Topic:          r.opts.TargetName,
		GroupID:        r.opts.GroupID,
		MinBytes:       r.opts.MinBytes,
		MaxBytes:       r.opts.MaxBytes,
		MaxWait:        r.opts.MaxWait,
		CommitInterval: 0,
		Logger:         r.kafkaLogger.InfoLogger(),
		ErrorLogger:    r.kafkaLogger.ErrorLogger(),
	})

	r.mu.Lock()
	r.reader = newReader
	r.mu.Unlock()

	return newReader, nil
}
