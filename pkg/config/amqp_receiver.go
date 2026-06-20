package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type AMQPReceiverConfig struct {
	URL        string `mapstructure:"url" json:"url,omitempty" yaml:"url,omitempty"`                         // Хост и порт брокера
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"` // Имя очереди, которую слушаем
	Username   string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password   string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`

	// Настройки безопасности
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty"` // Пропуск валидации сертификатов (для local dev)

	// Сетевые таймауты получателя
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`    // Default: 5s
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty"`             // Default: 30s
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Default: 5s

	// Параметры Backpressure для Receiver / Consumer
	PrefetchCredit  int  `mapstructure:"prefetch_credit" json:"prefetch_credit,omitempty" yaml:"prefetch_credit,omitempty"`    // Default: 100
	WorkerCount     int  `mapstructure:"worker_count" json:"worker_count,omitempty" yaml:"worker_count,omitempty"`             // Default: 10
	DataCapacity    int  `mapstructure:"data_capacity" json:"data_capacity,omitempty" yaml:"data_capacity,omitempty"`          // Default: 200
	CompleteProcess bool `mapstructure:"complete_process" json:"complete_process,omitempty" yaml:"complete_process,omitempty"` // Default: true
}

func NewAMQPReceiverConfig(
	url string,
	targetName string,
	username string,
	password string,
	insecureSkipVerify bool,
	connectTimeout time.Duration,
	idleTimeout time.Duration,
	shutdownTimeout time.Duration,
	prefetchCredit int,
	workerCount int,
	dataCapacity int,
	completeProcess bool,
) *AMQPReceiverConfig {
	return &AMQPReceiverConfig{
		URL:                url,
		TargetName:         targetName,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecureSkipVerify,
		ConnectTimeout:     connectTimeout,
		IdleTimeout:        idleTimeout,
		ShutdownTimeout:    shutdownTimeout,
		PrefetchCredit:     prefetchCredit,
		WorkerCount:        workerCount,
		DataCapacity:       dataCapacity,
		CompleteProcess:    completeProcess,
	}
}

func NewDefaultAMQPReceiverConfig() *AMQPReceiverConfig {
	return NewAMQPReceiverConfig(
		DefaultAMQPReceiverURL,
		"",
		"",
		"",
		DefaultAMQPReceiverInsecureSkipVerify,
		DefaultAMQPReceiverConnectTimeout,
		DefaultAMQPReceiverIdleTimeout,
		DefaultAMQPReceiverShutdownTimeout,
		DefaultAMQPReceiverPrefetchCredit,
		DefaultAMQPReceiverWorkerCount,
		DefaultAMQPReceiverDataCapacity,
		DefaultAMQPReceiverCompleteProcess,
	)
}

func (rc *AMQPReceiverConfig) Validate() error {
	if strings.TrimSpace(rc.URL) == "" {
		return errs.NewConfigValidateError("amqp receiver", "URL", "empty", nil)
	}
	if strings.TrimSpace(rc.TargetName) == "" {
		return errs.NewConfigValidateError("amqp receiver", "TargetName", "empty", nil)
	}
	if !(rc.ConnectTimeout > 0) {
		return errs.NewConfigValidateError("amqp receiver", "ConnectTimeout", "less than 0", nil)
	}
	if !(rc.IdleTimeout > 0) {
		return errs.NewConfigValidateError("amqp receiver", "IdleTimeout", "less than 0", nil)
	}
	if !(rc.ShutdownTimeout > 0) {
		return errs.NewConfigValidateError("amqp receiver", "ShutdownTimeout", "less than 0", nil)
	}
	if !(rc.PrefetchCredit > 0) {
		return errs.NewConfigValidateError("amqp receiver", "PrefetchCredit", "less than 0", nil)
	}
	if !(rc.WorkerCount > 0) {
		return errs.NewConfigValidateError("amqp receiver", "WorkerCount", "less than 0", nil)
	}
	if !(rc.DataCapacity > 0) {
		return errs.NewConfigValidateError("amqp receiver", "DataCapacity", "less than 0", nil)
	}

	return nil
}
