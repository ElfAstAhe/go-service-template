package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AMQPSenderConfig currently usable for artemis amq, later improve for artemis,kafka,rabbit
type AMQPSenderConfig struct {
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"` // Имя очереди (FQQN)

	// Сетевые таймауты отправителя (Prod Way)
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`    // Default: 5s
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Default: 3s

	// publish
	PublishMaxTryAttempts int           `mapstructure:"publish_max_try_attempts" json:"publish_max_try_attempts" yaml:"publish_max_try_attempts"`
	PublishBaseRetryDelay time.Duration `mapstructure:"publish_base_retry_delay" json:"publish_base_retry_delay,omitempty" yaml:"publish_base_retry_delay,omitempty"`
	PublishMaxRetryDelay  time.Duration `mapstructure:"publish_max_retry_delay" json:"publish_max_retry_delay" yaml:"publish_max_retry_delay"`
}

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

func NewDefaultAMQPSenderConfig() *AMQPSenderConfig {
	return NewAMQPSenderConfig(
		"",
		DefaultAMQPSenderConnectTimeout,
		DefaultAMQPSenderShutdownTimeout,
		DefaultAMQPSenderPublishMaxTryAttempts,
		DefaultAMQPSenderPublishBaseRetryDelay,
		DefaultAMQPSenderPublishMaxRetryDelay,
	)
}

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
	if sc.PublishBaseRetryDelay > sc.PublishMaxRetryDelay {
		return errs.NewConfigValidateError("amqp sender", "PublishMaxRetryDelay", "less than base delay", nil)
	}

	return nil
}
