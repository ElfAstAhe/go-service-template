package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type AMQPSenderConfig struct {
	URL        string `mapstructure:"url" json:"url,omitempty" yaml:"url,omitempty"`                         // Хост и порт брокера (например, "localhost:5672")
	Address    string `mapstructure:"address" json:"address,omitempty" yaml:"address,omitempty"`             // Адрес очереди/топика
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"` // Имя очереди, которую слушаем
	Username   string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password   string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`

	// Настройки безопасности
	Secure             bool `mapstructure:"secure" json:"secure,omitempty" yaml:"secure,omitempty"`                                           // Включает шифрование трафика (SSL/TLS)
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty"` // Пропуск валидации сертификатов (для local dev)

	// Сетевые таймауты отправителя (Prod Way)
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"` // Default: 5s
	WriteTimeout   time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`       // Default: 3s

	// Таймауты оркестрации отправки шаблона
	NotifyTimeout   time.Duration `mapstructure:"notify_timeout" json:"notify_timeout,omitempty" yaml:"notify_timeout,omitempty"`       // Default: 2s
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"` // Default: 3s
}

func NewAMQPSenderConfig(
	url string,
	address string,
	targetName string,
	username string,
	password string,
	secure bool,
	insecureSkipVerify bool,
	connectTimeout time.Duration,
	writeTimeout time.Duration,
	notifyTimeout time.Duration,
	shutdownTimeout time.Duration,
) *AMQPSenderConfig {
	return &AMQPSenderConfig{
		URL:                url,
		Address:            address,
		TargetName:         targetName,
		Username:           username,
		Password:           password,
		Secure:             secure,
		InsecureSkipVerify: insecureSkipVerify,
		ConnectTimeout:     connectTimeout,
		WriteTimeout:       writeTimeout,
		NotifyTimeout:      notifyTimeout,
		ShutdownTimeout:    shutdownTimeout,
	}
}

func NewDefaultAMQPSenderConfig() *AMQPSenderConfig {
	return NewAMQPSenderConfig(
		DefaultAMQPSenderURL,
		"",
		"",
		"",
		"",
		DefaultAMQPSenderSecure,
		DefaultAMQPSenderInsecureSkipVerify,
		DefaultAMQPSenderConnectTimeout,
		DefaultAMQPSenderWriteTimeout,
		DefaultAMQPSenderNotifyTimeout,
		DefaultAMQPSenderShutdownTimeout,
	)
}

func (sc *AMQPSenderConfig) Validate() error {
	if strings.TrimSpace(sc.URL) == "" {
		return errs.NewConfigValidateError("amqp sender", "URL", "empty", nil)
	}
	if strings.TrimSpace(sc.Address) == "" {
		return errs.NewConfigValidateError("amqp sender", "Address", "empty", nil)
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

	return nil
}
