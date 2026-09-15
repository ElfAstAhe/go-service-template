package kafka

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
)

type Sender struct {
	opts   *SenderOptions
	writer KafkaSenderLink
	logger logger.Logger
	mu     sync.RWMutex
	initMu sync.Mutex
}

// Привязываем к вашему дженерик-интерфейсу с типом опций any
var _ pkgamqp.Sender[any] = (*Sender)(nil)

func NewSender(opts ...SenderOption) (*Sender, error) {
	clientOpts := NewSenderOptions()
	for _, opt := range opts {
		opt(clientOpts)
	}
	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewSender", "kafka sender options validation failed", err)
	}
	return &Sender{
		opts:   clientOpts,
		logger: clientOpts.Logger.GetLogger("kafka-sender"),
	}, nil
}

func (s *Sender) Publish(ctx context.Context, msg pkgamqp.Message, _ any) error {
	if utils.IsNil(msg) {
		return errs.NewTlCommonError("Publish", "cannot publish nil message", nil)
	}

	for attempt := 1; attempt <= s.opts.PublishMaxTryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return errs.NewTlCommonError("Publish", "context canceled", err)
		}

		kafkaWriter, err := s.getSender(ctx)
		if err != nil {
			if attempt < s.opts.PublishMaxTryAttempts {
				s.waitBackoff(ctx, attempt)
				continue
			}
			return errs.NewTlCommonError("Publish", "kafka sender failed to init writer", err)
		}

		kafkaMsg := s.prepareMessage(msg)
		err = kafkaWriter.WriteMessages(ctx, kafkaMsg)
		if err == nil {
			return nil
		}

		err = s.handleSendError(attempt, err)
		if err != nil {
			return err
		}

		s.waitBackoff(ctx, attempt)
	}
	return errs.NewTlCommonError("Publish", "kafka sender unexpected retry loop exit", nil)
}

func (s *Sender) Close(ctx context.Context) error {
	s.mu.Lock()
	writerToClose := s.writer
	s.writer = nil
	s.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)
	go func() {
		var closeErrs []error
		if !utils.IsNil(writerToClose) {
			if err := writerToClose.Close(); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}
		done <- errors.Join(closeErrs...)
	}()

	select {
	case err := <-done:
		if err != nil {
			return errs.NewTlCommonError("Close", "kafka sender close fails", err)
		}
		return nil
	case <-closeCtx.Done():
		return errs.NewTlCommonError("Close", "kafka sender close timeout", closeCtx.Err())
	}
}

func (s *Sender) GetTargetName() string { return s.opts.TargetName }

func (s *Sender) getSender(ctx context.Context) (KafkaSenderLink, error) {
	s.mu.RLock()
	if !utils.IsNil(s.writer) {
		defer s.mu.RUnlock()
		return s.writer, nil
	}
	s.mu.RUnlock()

	s.initMu.Lock()
	defer s.initMu.Unlock()

	s.mu.RLock()
	if !utils.IsNil(s.writer) {
		defer s.mu.RUnlock()
		return s.writer, nil
	}
	s.mu.RUnlock()

	_, err := s.opts.Connector.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	// Приведение интерфейса к конкретному типу Connector для извлечения настроек
	localConn, ok := s.opts.Connector.(*Connector)
	if !ok {
		return nil, fmt.Errorf("invalid connector type, expected *kafka.Connector")
	}

	newWriter := &kafka.Writer{
		Addr:         kafka.TCP(localConn.GetBrokers()...),
		Topic:        s.opts.TargetName,
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  1, // Управление повторами на стороне нашего метода Publish
		WriteTimeout: s.opts.ConnectTimeout,
	}

	s.mu.Lock()
	s.writer = newWriter
	s.mu.Unlock()
	return newWriter, nil
}

func (s *Sender) handleSendError(attempt int, err error) error {
	var netErr net.Error
	var kErr kafka.Error
	if errors.As(err, &netErr) || errors.As(err, &kErr) {
		if attempt < s.opts.PublishMaxTryAttempts {
			s.opts.Connector.Invalidate(err)
			s.mu.Lock()
			s.writer = nil
			s.mu.Unlock()
			return nil
		}
		return errs.NewTlCommonError("Publish", "kafka sender network error persisted", err)
	}
	return errs.NewTlCommonError("Publish", "kafka unrecoverable send error", err)
}

func (s *Sender) waitBackoff(ctx context.Context, attempt int) {
	shift := min(uint(attempt-1), 31)
	delay := s.opts.PublishBaseRetryDelay * (1 << shift)
	if delay > s.opts.PublishMaxRetryDelay || delay <= 0 {
		delay = s.opts.PublishMaxRetryDelay
	}
	delayMs := int(delay / time.Millisecond)

	if delayMs > 5 {
		maxJitterMs := delayMs / 5
		jitterMs := rand.IntN(maxJitterMs)
		jitter := time.Duration(jitterMs) * time.Millisecond
		if rand.IntN(2) == 0 {
			delay += jitter
		} else {
			delay -= jitter
		}
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (s *Sender) prepareMessage(msg pkgamqp.Message) kafka.Message {
	kafkaMsg := kafka.Message{Value: msg.GetPayload()}
	if len(msg.GetProperties()) > 0 {
		if keyVal, ok := msg.GetProperties()["kafka_message_key"]; ok {
			if strKey, ok := keyVal.(string); ok {
				kafkaMsg.Key = []byte(strKey)
			}
		}
		var headers []kafka.Header
		for k, v := range msg.GetProperties() {
			if k == "kafka_message_key" {
				continue
			}
			if strVal, ok := v.(string); ok {
				headers = append(headers, kafka.Header{Key: k, Value: []byte(strVal)})
			}
		}
		kafkaMsg.Headers = headers
	}
	return kafkaMsg
}
