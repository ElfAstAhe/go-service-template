package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
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
	if !(sc.ConnectTimeout > 0) {
		return errs.NewConfigValidateError("amqp sender", "ConnectTimeout", "less than 0", nil)
	}
	if !(sc.ShutdownTimeout > 0) {
		return errs.NewConfigValidateError("amqp sender", "ShutdownTimeout", "less than 0", nil)
	}
	if !(sc.PublishMaxTryAttempts > 1) {
		return errs.NewConfigValidateError("amqp sender", "PublishMaxTryAttempts", "less than 1", nil)
	}
	if !(sc.PublishBaseRetryDelay > 0) {
		return errs.NewConfigValidateError("amqp sender", "PublishBaseRetryDelay", "less than 0", nil)
	}
	if !(sc.PublishMaxRetryDelay > 0) {
		return errs.NewConfigValidateError("amqp sender", "PublishMaxRetryDelay", "less than 0", nil)
	}
	// Валидация математической консистентности: базовый шаг не может превышать максимальный потолок
	if sc.PublishBaseRetryDelay > sc.PublishMaxRetryDelay {
		return errs.NewConfigValidateError("amqp sender", "PublishMaxRetryDelay", "less than base delay", nil)
	}

	return nil
}
