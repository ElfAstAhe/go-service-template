package azure

import (
	"strings"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
)

// Константы рантайм-дефолтов для внутренней защиты отправителя Azure Service Bus.
const (
	DefaultSenderConnectTimeout        time.Duration = 5 * time.Second
	DefaultSenderShutdownTimeout       time.Duration = 5 * time.Second
	DefaultSenderPublishMaxTryAttempts int           = 2
	DefaultSenderPublishBaseRetryDelay time.Duration = 100 * time.Millisecond
	DefaultSenderPublishMaxRetryDelay  time.Duration = 3 * time.Second
)

// SenderOption определяет функциональный тип для конфигурации опций отправителя (Fluent API).
type SenderOption func(*SenderOptions)

// SenderOptions инкапсулирует все параметры рантайма, необходимые для безопасной сборки и работы Sender.
type SenderOptions struct {
	// Connector представляет собой ссылку на глобальный дженерик-коннектор для управления AMQP-сессиями.
	Connector broker.Connector[*amqp.Session]
	// TargetName определяет точное имя целевой очереди (Queue) или топика (Topic) в Azure Service Bus.
	TargetName string
	// Opts содержит специфичные низкоуровневые настройки линка отправки библиотеки Azure go-amqp.
	Opts *amqp.SenderOptions
	// SendOpts инкапсулирует дефолтные параметры для единичных вызовов публикации (используется плоским Publish).
	SendOpts        *amqp.SendOptions
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	Logger          logger.Logger
	// PublishMaxTryAttempts задает верхний лимит повторных попыток отправки при временных сетевых сбоях.
	PublishMaxTryAttempts int
	// PublishBaseRetryDelay определяет базовый стартовый интервал для формулы экспоненциального бэкоффа.
	PublishBaseRetryDelay time.Duration
	// PublishMaxRetryDelay ограничивает максимальную задержку между попытками публикации (жесткий потолок).
	PublishMaxRetryDelay time.Duration
}

// NewSenderOptions создает структуру опций, сразу наполняя её безопасными базовыми таймаутами.
func NewSenderOptions() *SenderOptions {
	return &SenderOptions{
		ConnectTimeout:        DefaultSenderConnectTimeout,
		ShutdownTimeout:       DefaultSenderShutdownTimeout,
		PublishMaxTryAttempts: DefaultSenderPublishMaxTryAttempts,
		PublishBaseRetryDelay: DefaultSenderPublishBaseRetryDelay,
		PublishMaxRetryDelay:  DefaultSenderPublishMaxRetryDelay,
	}
}

// Validate выполняет строгую проверку входящих параметров рантайма перед созданием Sender.
// Защищает приложение от логических ошибок конфигурации бэкоффа и паник при отсутствии зависимостей.
func (so *SenderOptions) Validate() error {
	if so.Connector == nil {
		return errs.NewTlCommonError("Validate", "connector is required and cannot be nil", nil)
	}
	if strings.TrimSpace(so.TargetName) == "" {
		return errs.NewTlCommonError("Validate", "target name empty", nil)
	}
	if so.ConnectTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "connection timeout is invalid", nil)
	}
	if so.ShutdownTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "shutdown timeout is invalid", nil)
	}
	if so.PublishMaxTryAttempts <= 0 {
		return errs.NewTlCommonError("Validate", "publish max try attempts is invalid", nil)
	}
	if so.PublishBaseRetryDelay <= 0 {
		return errs.NewTlCommonError("Validate", "publish base retry delay is invalid", nil)
	}
	if so.PublishMaxRetryDelay <= 0 {
		return errs.NewTlCommonError("Validate", "publish max retry delay is invalid", nil)
	}
	if so.Logger == nil {
		return errs.NewTlCommonError("Validate", "logger is nil", nil)
	}
	// Золотое правило экспоненциального бэкоффа: стартовая задержка не может превышать максимальный лимит
	if so.PublishBaseRetryDelay > so.PublishMaxRetryDelay {
		return errs.NewTlCommonError("Validate", "base retry delay cannot be greater than max retry delay", nil)
	}

	return nil
}

// ====================================================================
// Fluent API методы конфигурации
// ====================================================================

// WithSenderConnector инжектирует ссылку на глобальный менеджер AMQP-сессий коннектора.
func WithSenderConnector(connector broker.Connector[*amqp.Session]) SenderOption {
	return func(so *SenderOptions) { so.Connector = connector }
}

// WithSenderTargetName задает имя целевой очереди или топика назначения для публикации сообщений.
func WithSenderTargetName(targetName string) SenderOption {
	return func(so *SenderOptions) { so.TargetName = targetName }
}

// WithSenderConnectTimeout переопределяет время ожидания при создании сетевого линка отправки.
func WithSenderConnectTimeout(timeout time.Duration) SenderOption {
	return func(so *SenderOptions) { so.ConnectTimeout = timeout }
}

// WithSenderShutdownTimeout переопределяет таймаут для мягкого гашения и сброса буферов линка продюсера.
func WithSenderShutdownTimeout(timeout time.Duration) SenderOption {
	return func(so *SenderOptions) { so.ShutdownTimeout = timeout }
}

// WithSenderLogger инжектирует структурированный логгер для аудита операций публикации сообщений.
func WithSenderLogger(log logger.Logger) SenderOption {
	return func(so *SenderOptions) { so.Logger = log }
}

// WithSenderPublishMaxTryAttempts настраивает верхний лимит повторных попыток публикации (Retries).
func WithSenderPublishMaxTryAttempts(maxTryAttempts int) SenderOption {
	return func(so *SenderOptions) { so.PublishMaxTryAttempts = maxTryAttempts }
}

// WithSenderPublishBaseRetryDelay задает стартовый шаг временной задержки экспоненциального отката.
func WithSenderPublishBaseRetryDelay(delay time.Duration) SenderOption {
	return func(so *SenderOptions) { so.PublishBaseRetryDelay = delay }
}

// WithSenderPublishMaxRetryDelay фиксирует жесткий потолок задержки ретрая для предотвращения бесконечного роста бэкоффа.
func WithSenderPublishMaxRetryDelay(delay time.Duration) SenderOption {
	return func(so *SenderOptions) { so.PublishMaxRetryDelay = delay }
}

// WithSenderOpts позволяет передать низкоуровневые параметры конфигурации линка отправки библиотеки Azure go-amqp.
func WithSenderOpts(senderOpts *amqp.SenderOptions) SenderOption {
	return func(so *SenderOptions) { so.Opts = senderOpts }
}

// WithSendOpts настраивает базовые опции доставки единичных сообщений (например, флаг Settled).
func WithSendOpts(opts *amqp.SendOptions) SenderOption {
	return func(so *SenderOptions) { so.SendOpts = opts }
}
