package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AMQPReceiverConfig инкапсулирует конфигурационные параметры получателя (Consumer / Receiver)
// для чтения сообщений из брокеров по протоколу AMQP 0-9-1.
//
// Управляет как сетевыми лимитами рантайма, так и стратегиями Backpressure для защиты
// оперативной памяти приложения от лавинообразных нагрузок.
type AMQPReceiverConfig struct {
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"` // Имя очереди брокера, которую слушает текущий воркер

	// Сетевые таймауты получателя
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`    // Лимит времени на установку логического канала (Default: 5s)
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Лимит времени на мягкую остановку и коммит offset-ов (Default: 5s)

	// Параметры Backpressure для Receiver / Consumer
	PrefetchCredit int `mapstructure:"prefetch_credit" json:"prefetch_credit,omitempty" yaml:"prefetch_credit,omitempty"` // Лимит QoS Prefetch: сколько сообщений воркер может удерживать в памяти без Ack (Default: 100)
}

// NewAMQPReceiverConfig — фабричный конструктор конфигурации AMQP получателя.
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

// NewDefaultAMQPReceiverConfig собирает конфигурацию по умолчанию, наполняя её системными константными таймаутами.
func NewDefaultAMQPReceiverConfig() *AMQPReceiverConfig {
	return NewAMQPReceiverConfig(
		"", // Намеренно оставляем пустым, требуя явного переопределения в конфигурационных файлах
		DefaultAMQPReceiverConnectTimeout,
		DefaultAMQPReceiverShutdownTimeout,
		DefaultAMQPReceiverPrefetchCredit,
	)
}

// Validate осуществляет семантическую проверку параметров конфигурации на этапе запуска приложения (Bootstrap Phase).
// Полностью пресекает попытки запуска консьюмера брокера с отрицательным окном Backpressure или пустым именем очереди.
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
