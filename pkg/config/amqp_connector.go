package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AMQPConnectorConfig connector configuration
type AMQPConnectorConfig struct {
	// host,port
	URL string `mapstructure:"url" json:"url,omitempty" yaml:"url,omitempty"` // Хост и порт брокера (например, "localhost:5672")
	// security
	Username string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`

	// Сетевые таймауты отправителя (Prod Way)
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`    // Default: 5s
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty"`             // Default: 30s
	WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`          // Default: 3s
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Default: 3s
}

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

func NewDefaultAMQPConnectorConfig() *AMQPConnectorConfig {
	return NewAMQPConnectorConfig(
		DefaultAMQPConnectorURL,
		"",
		"",
		DefaultAMQPConnectorConnectTimeout,
		DefaultAMQPConnectorShutdownTimeout,
		DefaultAMQPConnectorWriteTimeout,
		DefaultAMQPConnectorIdleTimeout,
	)
}

func (cc *AMQPConnectorConfig) Validate() error {
	if strings.TrimSpace(cc.URL) == "" {
		return errs.NewConfigValidateError("amqp connector", "URL", "empty", nil)
	}
	if !(cc.ConnectTimeout > 0) {
		return errs.NewConfigValidateError("amqp connector", "ConnectTimeout", "less than 0", nil)
	}
	if !(cc.ShutdownTimeout > 0) {
		return errs.NewConfigValidateError("amqp connector", "ShutdownTimeout", "less than 0", nil)
	}
	if !(cc.WriteTimeout > 0) {
		return errs.NewConfigValidateError("amqp connector", "WriteTimeout", "less than 0", nil)
	}
	if !(cc.IdleTimeout > 0) {
		return errs.NewConfigValidateError("amqp connector", "IdleTimeout", "less than 0", nil)
	}

	return nil
}
