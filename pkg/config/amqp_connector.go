package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// ====================================================================
// amqp connector
// ====================================================================
// Настройки по умолчанию для сетевого пула соединений брокеров сообщений по протоколу AMQP 0-9-1.
const (
	DefaultAMQPConnectorURL             string        = "amqp://localhost:5672/" // Сетевой адрес брокера RabbitMQ по умолчанию
	DefaultAMQPConnectorConnectTimeout  time.Duration = 10 * time.Second         // Лимит времени на установку TCP-сессии
	DefaultAMQPConnectorWriteTimeout    time.Duration = 10 * time.Second         // Лимит времени на запись фрейма в брокер
	DefaultAMQPConnectorIdleTimeout     time.Duration = 30 * time.Second         // Лимит удержания простаивающего канала со стороны ОС
	DefaultAMQPConnectorShutdownTimeout time.Duration = 15 * time.Second         // Лимит времени на мягкое освобождение ресурсов канала
)

// AMQPConnectorConfig connector configuration
//
// Инкапсулирует конфигурационные параметры низкоуровневого сетевого подключения (Connector)
// к брокерам сообщений по протоколу AMQP 0-9-1 (RabbitMQ). Вынесение операционных таймаутов
// в дельту конфигурации является обязательным требованием для обеспечения отказоустойчивости.
type AMQPConnectorConfig struct {
	// host,port
	URL string `mapstructure:"url" json:"url,omitempty" yaml:"url,omitempty"` // Хост и порт брокера (например, "localhost:5672")
	// security
	Username string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`

	// Сетевые таймауты отправителя (Prod Way)
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty" toml:"connect_timeout,omitempty"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty" toml:"idle_timeout,omitempty"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty" toml:"write_timeout,omitempty"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty" toml:"shutdown_timeout,omitempty"`
}

// NewAMQPConnectorConfig конструктор
func NewAMQPConnectorConfig(
	url string,
	username string,
	password string,
	connectTimeout time.Duration,
	IdleTimeout time.Duration,
	writeTimeout time.Duration,
	shutdownTimeout time.Duration,
) *AMQPConnectorConfig {
	return &AMQPConnectorConfig{
		URL:             url,
		Username:        username,
		Password:        password,
		ConnectTimeout:  connectTimeout,
		ShutdownTimeout: shutdownTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     IdleTimeout,
	}
}

// NewDefaultAMQPConnectorConfig конструктор default настроек
func NewDefaultAMQPConnectorConfig() *AMQPConnectorConfig {
	return NewAMQPConnectorConfig(
		DefaultAMQPConnectorURL,
		"",
		"",
		DefaultAMQPConnectorConnectTimeout,
		DefaultAMQPConnectorWriteTimeout,
		DefaultAMQPConnectorIdleTimeout,
		DefaultAMQPConnectorShutdownTimeout,
	)
}

// Validate осуществляет семантическую проверку параметров конфигурации на этапе стартапа (Bootstrap Phase).
// Защищает рантайм от запуска коннектора с некорректными, пустыми или отрицательными сетевыми таймаутами.
func (cc *AMQPConnectorConfig) Validate() error {
	if strings.TrimSpace(cc.URL) == "" {
		return errs.NewConfigValidateError("amqp connector", "URL", "empty", nil)
	}
	if cc.ConnectTimeout <= 0 {
		return errs.NewConfigValidateError("amqp connector", "ConnectTimeout", "less than 0", nil)
	}
	if cc.ShutdownTimeout <= 0 {
		return errs.NewConfigValidateError("amqp connector", "ShutdownTimeout", "less than 0", nil)
	}
	if cc.WriteTimeout <= 0 {
		return errs.NewConfigValidateError("amqp connector", "WriteTimeout", "less than 0", nil)
	}
	if cc.IdleTimeout <= 0 {
		return errs.NewConfigValidateError("amqp connector", "IdleTimeout", "less than 0", nil)
	}

	return nil
}
