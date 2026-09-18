package kafka

import (
	"context"
	"errors"
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

// Sender реализует отправку сообщений в конкретный топик Kafka.
// Поддерживает политики экспоненциального бэкоффа с джиттером.
type Sender struct {
	opts        *SenderOptions
	writer      KafkaSenderLink
	logger      logger.Logger
	kafkaLogger *logger.KafkaLogger
	mu          sync.RWMutex
	initMu      sync.Mutex
}

// Привязываем структуру к общему интерфейсу amqp.Sender.
var _ pkgamqp.Sender[any] = (*Sender)(nil)

// NewSender создает новый экземпляр отправителя на основе переданных опций.
func NewSender(opts ...SenderOption) (*Sender, error) {
	clientOpts := NewSenderOptions()
	for _, opt := range opts {
		opt(clientOpts)
	}
	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewSender", "kafka sender options validation failed", err)
	}

	log := clientOpts.Logger.GetLogger("kafka-sender")

	return &Sender{
		opts:        clientOpts,
		logger:      log,
		kafkaLogger: logger.NewKafkaLogger(log),
	}, nil
}

// Publish отправляет сообщение в Kafka. Включает механизм повторных попыток при сетевых сбоях.
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
			return nil // Успешная отправка
		}

		// Логируем ошибку и проверяем, имеет ли смысл делать ретрай
		if !s.isRecoverableError(err) || attempt == s.opts.PublishMaxTryAttempts {
			return errs.NewTlCommonError("Publish", "kafka unrecoverable send error or retries exhausted", err)
		}

		s.logger.Warnf("Temporary error sending to Kafka (attempt %d/%d): %v. Retrying...", attempt, s.opts.PublishMaxTryAttempts, err)
		s.waitBackoff(ctx, attempt)
	}

	return errs.NewTlCommonError("Publish", "kafka sender unexpected retry loop exit", nil)
}

// Close плавно завершает работу врайтера, дожидаясь отправки пакетов из буферов.
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

// GetTargetName возвращает имя топика назначения.
func (s *Sender) GetTargetName() string {
	return s.opts.TargetName
}

// getSender инициализирует или возвращает существующий линк врайтера (Double-Checked Locking паттерн).
//
//goland:noinspection DuplicatedCode
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

	// Напрямую инициализируем современный потокобезопасный пуллер kafka.Writer
	newWriter := &kafka.Writer{
		Addr:         kafka.TCP(s.opts.Brokers...),
		Topic:        s.opts.TargetName,
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  1, // Повторами управляем сами в методе Publish с кастомным бэкоффом
		WriteTimeout: s.opts.ConnectTimeout,
		Logger:       s.kafkaLogger.InfoLogger(),
		ErrorLogger:  s.kafkaLogger.ErrorLogger(),
	}

	s.mu.Lock()
	s.writer = newWriter
	s.mu.Unlock()

	return newWriter, nil
}

// isRecoverableError определяет, является ли ошибка временной (сетевой), допуская повторную попытку.
func (s *Sender) isRecoverableError(err error) bool {
	var netErr net.Error
	var kErr kafka.Error
	return errors.As(err, &netErr) || errors.As(err, &kErr)
}

// waitBackoff вычисляет экспоненциальную задержку с добавлением случайного джиттера.
func (s *Sender) waitBackoff(ctx context.Context, attempt int) {
	shift := min(uint(attempt-1), 31)
	delay := s.opts.PublishBaseRetryDelay * (1 << shift)
	if delay > s.opts.PublishMaxRetryDelay || delay <= 0 {
		delay = s.opts.PublishMaxRetryDelay
	}
	delayMs := int(delay / time.Millisecond)

	if delayMs > 5 {
		maxJitterMs := delayMs / 5
		jitterMs := rand.IntN(maxJitterMs) // Использование потокобезопасного v2 крипто-рандомайзера
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

// prepareMessage перекладывает полезную нагрузку и свойства нашего конверта в нативную структуру kafka.Message.
func (s *Sender) prepareMessage(msg pkgamqp.Message) kafka.Message {
	kafkaMsg := kafka.Message{Value: msg.GetPayload()}
	props := msg.GetProperties()
	if len(props) > 0 {
		// Извлекаем ключ партиционирования Kafka
		if keyVal, ok := props["kafka_message_key"]; ok {
			if strKey, ok := keyVal.(string); ok {
				kafkaMsg.Key = []byte(strKey)
			}
		}

		// Перекладываем все остальные свойства в Headers
		var headers []kafka.Header
		for k, v := range props {
			if k == "kafka_message_key" || k == sysKafkaMsgKey {
				continue // Игнорируем служебные ключи нашего пакета
			}
			if strVal, ok := v.(string); ok {
				headers = append(headers, kafka.Header{Key: k, Value: []byte(strVal)})
			}
		}
		kafkaMsg.Headers = headers
	}
	return kafkaMsg
}
