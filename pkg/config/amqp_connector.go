package config

import (
	"time"
)

type AMQPConnectorConfig struct {
	URL string `mapstructure:"url" json:"url,omitempty" yaml:"url,omitempty"` // Хост и порт брокера (например, "localhost:5672")

	// Сетевые таймауты отправителя (Prod Way)
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`    // Default: 5s
	WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`          // Default: 3s
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Default: 3s
}
