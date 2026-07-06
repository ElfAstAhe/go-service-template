package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AMQPSenderConfig currently usable for artemis amq, later improve for artemis,kafka,rabbit
type AMQPSenderConfig struct {
	URL        string `mapstructure:"url" json:"url,omitempty" yaml:"url,omitempty"`                         // Хост и порт брокера (например, "localhost:5672")
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"` // Имя очереди (FQQN)
	Username   string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password   string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`

	// Настройки безопасности
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty"` // Пропуск валидации сертификатов (для local dev)

	// Сетевые таймауты отправителя (Prod Way)
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"` // Default: 5s
	WriteTimeout   time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`       // Default: 3s

	// Таймауты оркестрации отправки шаблона
	NotifyTimeout   time.Duration `mapstructure:"notify_timeout" json:"notify_timeout,omitempty" yaml:"notify_timeout,omitempty"`       // Default: 2s
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Default: 3s

	// publish
	PublishMaxTryAttempts int           `mapstructure:"publish_max_try_attempts" json:"publish_max_try_attempts" yaml:"publish_max_try_attempts"`
	PublishBaseRetryDelay time.Duration `mapstructure:"publish_base_retry_delay" json:"publish_base_retry_delay,omitempty" yaml:"publish_base_retry_delay,omitempty"`
	PublishMaxRetryDelay  time.Duration `mapstructure:"publish_max_retry_delay" json:"publish_max_retry_delay" yaml:"publish_max_retry_delay"`
}

func NewAMQPSenderConfig(
	url string,
	targetName string,
	username string,
	password string,
	insecureSkipVerify bool,
	connectTimeout time.Duration,
	writeTimeout time.Duration,
	notifyTimeout time.Duration,
	shutdownTimeout time.Duration,
	publishMaxTryAttempts int,
	publishBaseRetryDelay time.Duration,
	publishMaxRetryDelay time.Duration,
) *AMQPSenderConfig {
	return &AMQPSenderConfig{
		URL:                   url,
		TargetName:            targetName,
		Username:              username,
		Password:              password,
		InsecureSkipVerify:    insecureSkipVerify,
		ConnectTimeout:        connectTimeout,
		WriteTimeout:          writeTimeout,
		NotifyTimeout:         notifyTimeout,
		ShutdownTimeout:       shutdownTimeout,
		PublishMaxTryAttempts: publishMaxTryAttempts,
		PublishBaseRetryDelay: publishBaseRetryDelay,
		PublishMaxRetryDelay:  publishMaxRetryDelay,
	}
}

func NewDefaultAMQPSenderConfig() *AMQPSenderConfig {
	return NewAMQPSenderConfig(
		DefaultAMQPSenderURL,
		"",
		"",
		"",
		DefaultAMQPSenderInsecureSkipVerify,
		DefaultAMQPSenderConnectTimeout,
		DefaultAMQPSenderWriteTimeout,
		DefaultAMQPSenderNotifyTimeout,
		DefaultAMQPSenderShutdownTimeout,
		DefaultAMQPSenderPublishMaxTryAttempts,
		DefaultAMQPSenderPublishBaseRetryDelay,
		DefaultAMQPSenderPublishMaxRetryDelay,
	)
}

func (sc *AMQPSenderConfig) Validate() error {
	if strings.TrimSpace(sc.URL) == "" {
		return errs.NewConfigValidateError("amqp sender", "URL", "empty", nil)
	}
	if strings.TrimSpace(sc.TargetName) == "" {
		return errs.NewConfigValidateError("amqp sender", "TargetName", "empty", nil)
	}
	if !(sc.ConnectTimeout > 0) {
		return errs.NewConfigValidateError("amqp sender", "ConnectTimeout", "less than 0", nil)
	}
	if !(sc.WriteTimeout > 0) {
		return errs.NewConfigValidateError("amqp sender", "WriteTimeout", "less than 0", nil)
	}
	if !(sc.NotifyTimeout > 0) {
		return errs.NewConfigValidateError("amqp sender", "NotifyTimeout", "less than 0", nil)
	}
	if !(sc.ShutdownTimeout > 0) {
		return errs.NewConfigValidateError("amqp sender", "ShutdownTimeout", "less than 0", nil)
	}
	if !(sc.PublishBaseRetryDelay > 0) {
		return errs.NewConfigValidateError("amqp sender", "PublishBaseRetryDelay", "less than 0", nil)
	}
	if !(sc.PublishMaxRetryDelay > 0) {
		return errs.NewConfigValidateError("amqp sender", "PublishMaxRetryDelay", "less than 0", nil)
	}

	return nil
}
