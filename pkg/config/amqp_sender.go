package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// flags templates
const (
	// TemplateFlagAMQPSenderTargetName Имя целевой очереди или адреса брокера
	TemplateFlagAMQPSenderTargetName string = "%s-amqp-sender-target-name"
	// TemplateFlagAMQPSenderConnectTimeout Лимит времени на логическое подключение
	TemplateFlagAMQPSenderConnectTimeout string = "%s-amqp-sender-connect-timeout"
	// TemplateFlagAMQPSenderShutdownTimeout Лимит времени на Flush буферов при закрытии сендера
	TemplateFlagAMQPSenderShutdownTimeout string = "%s-amqp-sender-shutdown-timeout"
	// TemplateFlagAMQPSenderPublishMaxTryAttempts Максимальное количество попыток публикации фрейма
	TemplateFlagAMQPSenderPublishMaxTryAttempts string = "%s-amqp-sender-publish-max-try-attempts"
	// TemplateFlagAMQPSenderPublishBaseRetryDelay Стартовая базовая задержка повторной попытки (Exponential Backoff)
	TemplateFlagAMQPSenderPublishBaseRetryDelay string = "%s-amqp-sender-publish-base-retry-delay"
	// TemplateFlagAMQPSenderPublishMaxRetryDelay Верхний жесткий временной потолок задержки между попытками
	TemplateFlagAMQPSenderPublishMaxRetryDelay string = "%s-amqp-sender-publish-max-retry-delay"
)

// key templates
const (
	// TemplateKeyAMQPSenderTargetName Имя целевой очереди или адреса брокера
	TemplateKeyAMQPSenderTargetName string = "%s.target_name"
	// TemplateKeyAMQPSenderConnectTimeout Лимит времени на логическое подключение
	TemplateKeyAMQPSenderConnectTimeout string = "%s.connect_timeout"
	// TemplateKeyAMQPSenderShutdownTimeout Лимит времени на Flush буферов при закрытии сендера
	TemplateKeyAMQPSenderShutdownTimeout string = "%s.shutdown_timeout"
	// TemplateKeyAMQPSenderPublishMaxTryAttempts Максимальное количество попыток публикации фрейма
	TemplateKeyAMQPSenderPublishMaxTryAttempts string = "%s.publish_max_try_attempts"
	// TemplateKeyAMQPSenderPublishBaseRetryDelay Стартовая базовая задержка повторной попытки (Exponential Backoff)
	TemplateKeyAMQPSenderPublishBaseRetryDelay string = "%s.publish_base_retry_delay"
	// TemplateKeyAMQPSenderPublishMaxRetryDelay Верхний жесткий временной потолок задержки между попытками
	TemplateKeyAMQPSenderPublishMaxRetryDelay string = "%s.publish_max_retry_delay"
)

// ====================================================================
// amqp sender
// ====================================================================
// Параметры по умолчанию политики повторных попыток отправки фреймов (Publish Retry Policy) продюсера.
const (
	DefaultAMQPSenderConnectTimeout        time.Duration = 10 * time.Second       // Таймаут логического открытия продюсер-канала
	DefaultAMQPSenderShutdownTimeout       time.Duration = 15 * time.Second       // Таймаут на принудительный Flush буферов отправки
	DefaultAMQPSenderPublishMaxTryAttempts int           = 2                      // Лимит попыток публикации до возвращения критической ошибки
	DefaultAMQPSenderPublishBaseRetryDelay time.Duration = 100 * time.Millisecond // Стартовая задержка экспоненциального бэккоффа ретраев
	DefaultAMQPSenderPublishMaxRetryDelay  time.Duration = 3 * time.Second        // Верхний жесткий потолок ожидания между попытками
)

// AMQPSenderConfig инкапсулирует конфигурационные параметры отправителя сообщений (Producer / Publisher)
// для взаимодействия с очередями и брокерами платформы.
//
// Разработан с учетом поддержки FQQN (Fully Qualified Queue Name) для ActiveMQ Artemis,
// а также содержит расширенные настройки ретраев (Retry Policies) для обеспечения гарантированной доставки сообщений.
type AMQPSenderConfig struct {
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"` // Имя целевой очереди или адреса брокера (FQQN)

	// Сетевые таймауты отправителя (Prod Way конфигурация для предотвращения блокировок)
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`    // Лимит времени на логическое подключение (Default: 5s)
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Лимит времени на Flush буферов при закрытии сендера (Default: 3s)

	// Параметры политики повторных попыток отправки (Publish Retry Policy)
	PublishMaxTryAttempts int           `mapstructure:"publish_max_try_attempts" json:"publish_max_try_attempts" yaml:"publish_max_try_attempts"`                     // Максимальное количество попыток публикации фрейма
	PublishBaseRetryDelay time.Duration `mapstructure:"publish_base_retry_delay" json:"publish_base_retry_delay,omitempty" yaml:"publish_base_retry_delay,omitempty"` // Стартовая базовая задержка повторной попытки (Exponential Backoff)
	PublishMaxRetryDelay  time.Duration `mapstructure:"publish_max_retry_delay" json:"publish_max_retry_delay" yaml:"publish_max_retry_delay"`                        // Верхний жесткий временной потолок задержки между попытками
}

// NewAMQPSenderConfig — фабричный конструктор конфигурации AMQP отправителя.
func NewAMQPSenderConfig(
	targetName string,
	connectTimeout time.Duration,
	shutdownTimeout time.Duration,
	publishMaxTryAttempts int,
	publishBaseRetryDelay time.Duration,
	publishMaxRetryDelay time.Duration,
) *AMQPSenderConfig {
	return &AMQPSenderConfig{
		TargetName:            targetName,
		ConnectTimeout:        connectTimeout,
		ShutdownTimeout:       shutdownTimeout,
		PublishMaxTryAttempts: publishMaxTryAttempts,
		PublishBaseRetryDelay: publishBaseRetryDelay,
		PublishMaxRetryDelay:  publishMaxRetryDelay,
	}
}

// NewDefaultAMQPSenderConfig собирает конфигурацию по умолчанию, наполняя её системными константными таймаутами.
func NewDefaultAMQPSenderConfig() *AMQPSenderConfig {
	return NewAMQPSenderConfig(
		"", // Оставляем пустым, делегируя заполнение имени очереди (FQQN) конфигурационным файлам среды
		DefaultAMQPSenderConnectTimeout,
		DefaultAMQPSenderShutdownTimeout,
		DefaultAMQPSenderPublishMaxTryAttempts,
		DefaultAMQPSenderPublishBaseRetryDelay,
		DefaultAMQPSenderPublishMaxRetryDelay,
	)
}

// Validate осуществляет семантическую проверку параметров конфигурации на этапе запуска приложения (Bootstrap Phase).
// Предотвращает запуск продюсера с некорректными лимитами ретраев или математически неконсистентными интервалами задержек.
func (sc *AMQPSenderConfig) Validate() error {
	if strings.TrimSpace(sc.TargetName) == "" {
		return errs.NewConfigValidateError("amqp sender", "TargetName", "empty", nil)
	}
	if sc.ConnectTimeout <= 0 {
		return errs.NewConfigValidateError("amqp sender", "ConnectTimeout", "less than 0", nil)
	}
	if sc.ShutdownTimeout <= 0 {
		return errs.NewConfigValidateError("amqp sender", "ShutdownTimeout", "less than 0", nil)
	}
	if sc.PublishMaxTryAttempts < 2 {
		return errs.NewConfigValidateError("amqp sender", "PublishMaxTryAttempts", "less than 1", nil)
	}
	if sc.PublishBaseRetryDelay <= 0 {
		return errs.NewConfigValidateError("amqp sender", "PublishBaseRetryDelay", "less than 0", nil)
	}
	if sc.PublishMaxRetryDelay <= 0 {
		return errs.NewConfigValidateError("amqp sender", "PublishMaxRetryDelay", "less than 0", nil)
	}
	// Валидация математической консистентности: базовый шаг не может превышать максимальный потолок
	if sc.PublishBaseRetryDelay > sc.PublishMaxRetryDelay {
		return errs.NewConfigValidateError("amqp sender", "PublishMaxRetryDelay", "less than base delay", nil)
	}

	return nil
}

// DefaultAMQPSenderSettingsBuilder default AMQP Sender SettingsBuilder
func DefaultAMQPSenderSettingsBuilder(keyPrefix, flagPrefix string) *SettingsBuilder {
	return NewSettingsBuilder(keyPrefix, flagPrefix).
		WithTemplated(TemplateKeyAMQPSenderConnectTimeout, TemplateFlagAMQPSenderConnectTimeout, DefaultAMQPSenderConnectTimeout, "%s AMQP connect timeout").
		WithTemplated(TemplateKeyAMQPSenderShutdownTimeout, TemplateFlagAMQPSenderShutdownTimeout, DefaultAMQPSenderShutdownTimeout, "%s AMQP shutdown timeout").
		WithTemplated(TemplateKeyAMQPSenderPublishMaxTryAttempts, TemplateFlagAMQPSenderPublishMaxTryAttempts, DefaultAMQPSenderPublishMaxTryAttempts, "%s AMQP publish max send tries").
		WithTemplated(TemplateKeyAMQPSenderPublishBaseRetryDelay, TemplateFlagAMQPSenderPublishBaseRetryDelay, DefaultAMQPSenderPublishBaseRetryDelay, "%s AMQP publish base retry delay").
		WithTemplated(TemplateKeyAMQPSenderPublishMaxRetryDelay, TemplateFlagAMQPSenderPublishMaxRetryDelay, DefaultAMQPSenderPublishMaxRetryDelay, "%s AMQP publish max retry delay")
}
