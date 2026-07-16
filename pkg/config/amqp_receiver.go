package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type AMQPReceiverConfig struct {
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"` // Имя очереди, которую слушаем

	// Сетевые таймауты получателя
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`    // Default: 5s
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Default: 5s

	// Параметры Backpressure для Receiver / Consumer
	PrefetchCredit int `mapstructure:"prefetch_credit" json:"prefetch_credit,omitempty" yaml:"prefetch_credit,omitempty"` // Default: 100
}

func NewAMQPReceiverConfig(
	targetName string,
	connectTimeout time.Duration,
	shutdownTimeout time.Duration,
	prefetchCredit int,
) *AMQPReceiverConfig {
	return &AMQPReceiverConfig{
		TargetName:      targetName,
		ConnectTimeout:  connectTimeout,
		ShutdownTimeout: shutdownTimeout,
		PrefetchCredit:  prefetchCredit,
	}
}

func NewDefaultAMQPReceiverConfig() *AMQPReceiverConfig {
	return NewAMQPReceiverConfig(
		"",
		DefaultAMQPReceiverConnectTimeout,
		DefaultAMQPReceiverShutdownTimeout,
		DefaultAMQPReceiverPrefetchCredit,
	)
}

func (rc *AMQPReceiverConfig) Validate() error {
	if strings.TrimSpace(rc.TargetName) == "" {
		return errs.NewConfigValidateError("amqp receiver", "TargetName", "empty", nil)
	}
	if !(rc.ConnectTimeout > 0) {
		return errs.NewConfigValidateError("amqp receiver", "ConnectTimeout", "less than 0", nil)
	}
	if !(rc.ShutdownTimeout > 0) {
		return errs.NewConfigValidateError("amqp receiver", "ShutdownTimeout", "less than 0", nil)
	}
	if !(rc.PrefetchCredit > 0) {
		return errs.NewConfigValidateError("amqp receiver", "PrefetchCredit", "less than 0", nil)
	}

	return nil
}
