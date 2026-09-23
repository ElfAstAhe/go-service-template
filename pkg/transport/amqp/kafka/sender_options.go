package kafka

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/segmentio/kafka-go"
)

// Константы дефолтов для внутренней защиты рантайм-компонента (независимые от пакета config)
const (
	defaultSenderConnectTimeout              = 10 * time.Second
	defaultSenderIdleTimeout                 = 30 * time.Second
	defaultSenderShutdownTimeout             = 15 * time.Second
	defaultSenderPublishMaxTryAttempts       = 2
	defaultSenderPublishBaseRetryDelay       = 100 * time.Millisecond
	defaultSenderPublishMaxRetryDelay        = 3 * time.Second
	defaultSenderBatchSize                   = 100
	defaultSenderBatchBytes            int64 = 1048576
	defaultSenderBatchTimeout                = 10 * time.Millisecond
	defaultSenderWriteTimeout                = 10 * time.Second
	defaultSenderRequiredAcks                = kafka.RequiredAcks(kafka.RequireAll)
)

// Дефолтный адрес хоста для локальной разработки
var defaultSenderBrokers = []string{"localhost:9092"}

type SenderOption func(*SenderOptions)

// SenderOptions содержит параметры рантайма для сборки компонента Sender.
type SenderOptions struct {
	ClientID              string
	Brokers               []string
	TargetName            string
	WriterCustomizerFunc  func(*kafka.Writer)
	TLS                   *tls.Config
	ConnectTimeout        time.Duration
	IdleTimeout           time.Duration
	ShutdownTimeout       time.Duration
	PublishMaxTryAttempts int
	PublishBaseRetryDelay time.Duration
	PublishMaxRetryDelay  time.Duration
	Logger                logger.Logger
	Username              string
	Password              string
	BatchSize             int
	BatchBytes            int64
	BatchTimeout          time.Duration
	WriteTimeout          time.Duration
	RequiredAcks          kafka.RequiredAcks
}

// NewSenderOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewSenderOptions() *SenderOptions {
	return &SenderOptions{
		Brokers:               defaultSenderBrokers,
		ConnectTimeout:        defaultSenderConnectTimeout,
		IdleTimeout:           defaultSenderIdleTimeout,
		ShutdownTimeout:       defaultSenderShutdownTimeout,
		PublishMaxTryAttempts: defaultSenderPublishMaxTryAttempts,
		PublishBaseRetryDelay: defaultSenderPublishBaseRetryDelay,
		PublishMaxRetryDelay:  defaultSenderPublishMaxRetryDelay,
		BatchSize:             defaultSenderBatchSize,
		BatchBytes:            defaultSenderBatchBytes,
		BatchTimeout:          defaultSenderBatchTimeout,
		WriteTimeout:          defaultSenderWriteTimeout,
		RequiredAcks:          defaultSenderRequiredAcks,
	}
}

// Validate проверяет корректность абсолютно всех параметров рантайма отправителя.
// Гарантирует стабильность асинхронного батчинга и политик повторных попыток (Retry).
func (so *SenderOptions) Validate() error {
	// 1. Проверка базовых инфраструктурных полей
	if len(so.Brokers) == 0 {
		return errs.NewTlCommonError("Validate", "at least one broker address is required", nil)
	}
	if strings.TrimSpace(so.TargetName) == "" {
		return errs.NewTlCommonError("Validate", "target name (topic) cannot be empty", nil)
	}
	if so.Logger == nil {
		return errs.NewTlCommonError("Validate", "logger is required and cannot be nil", nil)
	}

	// 2. Проверка системных таймаутов жизненного цикла компонента
	if so.ConnectTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "ConnectTimeout must be greater than 0", nil)
	}
	if so.ShutdownTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "ShutdownTimeout must be greater than 0", nil)
	}

	// 3. Валидация политики повторных попыток отправки (Publish Retry)
	if so.PublishMaxTryAttempts < 1 {
		return errs.NewTlCommonError("Validate", "publish max try attempts must be at least 1", nil)
	}
	if so.PublishBaseRetryDelay <= 0 {
		return errs.NewTlCommonError("Validate", "PublishBaseRetryDelay must be greater than 0", nil)
	}
	if so.PublishMaxRetryDelay <= 0 {
		return errs.NewTlCommonError("Validate", "PublishMaxRetryDelay must be greater than 0", nil)
	}
	if so.PublishBaseRetryDelay > so.PublishMaxRetryDelay {
		return errs.NewTlCommonError("Validate", "PublishBaseRetryDelay cannot be greater than PublishMaxRetryDelay", nil)
	}

	// 4. Валидация параметров асинхронного пакетирования (Батчинга)
	if so.BatchSize <= 0 {
		return errs.NewTlCommonError("Validate", "BatchSize (messages pack limit) must be greater than 0", nil)
	}
	if so.BatchBytes <= 0 {
		return errs.NewTlCommonError("Validate", "BatchBytes (pack size in bytes) must be greater than 0", nil)
	}
	if so.BatchTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "BatchTimeout (flush period for incomplete pack) must be greater than 0", nil)
	}

	// 5. Валидация сетевого таймаута сокета на запись пачки
	if so.WriteTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "WriteTimeout (socket produce write deadline) must be greater than 0", nil)
	}

	// 6. Валидация политик подтверждения записи со стороны брокера
	// Согласно спецификации протокола Kafka:
	// -1 — all (весь ISR пул), 0 — пожар и забыл (без подтверждений), 1 — только лидер
	if so.RequiredAcks < -1 || so.RequiredAcks > 1 {
		return errs.NewTlCommonError("Validate", "RequiredAcks is invalid (must be -1 [all], 0 [none], or 1 [leader])", nil)
	}

	return nil
}

// ====================================================================
// Fluent API методы для сборки опций отправителя
// ====================================================================

func WithSenderClientID(clientID string) SenderOption {
	return func(so *SenderOptions) {
		so.ClientID = clientID
	}
}

func WithSenderBrokers(brokers []string) SenderOption {
	return func(so *SenderOptions) { so.Brokers = brokers }
}

func WithSenderTargetName(targetName string) SenderOption {
	return func(so *SenderOptions) { so.TargetName = targetName }
}

// WithSenderWriterCustomizer позволяет конечной системе тонко настроить любые специфичные поля kafka.Writer
func WithSenderWriterCustomizer(customizer func(*kafka.Writer)) SenderOption {
	return func(so *SenderOptions) {
		so.WriterCustomizerFunc = customizer
	}
}

func WithSenderTLS(tls *tls.Config) SenderOption {
	return func(so *SenderOptions) { so.TLS = tls }
}

func WithSenderConnectTimeout(timeout time.Duration) SenderOption {
	return func(so *SenderOptions) { so.ConnectTimeout = timeout }
}

func WithSenderIdleTimeout(timeout time.Duration) SenderOption {
	return func(so *SenderOptions) { so.IdleTimeout = timeout }
}

func WithSenderShutdownTimeout(timeout time.Duration) SenderOption {
	return func(so *SenderOptions) { so.ShutdownTimeout = timeout }
}

func WithSenderLogger(log logger.Logger) SenderOption {
	return func(so *SenderOptions) { so.Logger = log }
}

func WithSenderPublishRetry(maxAttempts int, baseDelay, maxDelay time.Duration) SenderOption {
	return func(so *SenderOptions) {
		so.PublishMaxTryAttempts = maxAttempts
		so.PublishBaseRetryDelay = baseDelay
		so.PublishMaxRetryDelay = maxDelay
	}
}

func WithSenderSecurity(username, password string) SenderOption {
	return func(so *SenderOptions) {
		so.Username = username
		so.Password = password
	}
}

func WithSenderBatchOptions(batchSize int, batchBytes int64, batchTimeout, writeTimeout time.Duration, requiredAcks kafka.RequiredAcks) SenderOption {
	return func(so *SenderOptions) {
		so.BatchSize = batchSize
		so.BatchBytes = batchBytes
		so.BatchTimeout = batchTimeout
		so.WriteTimeout = writeTimeout
		so.RequiredAcks = requiredAcks
	}
}
