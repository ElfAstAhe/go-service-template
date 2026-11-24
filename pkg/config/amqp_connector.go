package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// flags template
const (
	// TemplateFlagAMQPConnectorURL Сетевой адрес брокера AMQP v1.0
	TemplateFlagAMQPConnectorURL string = "%s-amqp-connector-url"
	// TemplateFlagAMQPConnectorUsername пользователь
	TemplateFlagAMQPConnectorUsername string = "%s-amqp-connector-username"
	// TemplateFlagAMQPConnectorPassword пароль
	//nolint:gosec // G101 : это флаги (параметры)
	TemplateFlagAMQPConnectorPassword string = "%s-amqp-connector-password"
	// TemplateFlagAMQPConnectorConnectTimeout Лимит времени на установку TCP-сессии
	TemplateFlagAMQPConnectorConnectTimeout string = "%s-amqp-connector-connect-timeout"
	// TemplateFlagAMQPConnectorIdleTimeout Лимит удержания простаивающего канала со стороны ОС
	TemplateFlagAMQPConnectorIdleTimeout string = "%s-amqp-connector-idle-timeout"
	// TemplateFlagAMQPConnectorWriteTimeout Лимит времени на запись фрейма в брокер
	TemplateFlagAMQPConnectorWriteTimeout string = "%s-amqp-connector-write-timeout"
	// TemplateFlagAMQPConnectorShutdownTimeout Лимит времени на мягкое освобождение ресурсов канала
	TemplateFlagAMQPConnectorShutdownTimeout string = "%s-amqp-connector-shutdown-timeout"
)

// keys template
const (
	// TemplateKeyAMQPConnectorURL Сетевой адрес брокера AMQP v1.0
	TemplateKeyAMQPConnectorURL string = "%s.url"
	// TemplateKeyAMQPConnectorUsername пользователь
	TemplateKeyAMQPConnectorUsername string = "%s.username"
	// TemplateKeyAMQPConnectorPassword пароль
	TemplateKeyAMQPConnectorPassword string = "%s.password"
	// TemplateKeyAMQPConnectorConnectTimeout Лимит времени на установку TCP-сессии
	TemplateKeyAMQPConnectorConnectTimeout string = "%s.connect_timeout"
	// TemplateKeyAMQPConnectorIdleTimeout Лимит удержания простаивающего канала со стороны ОС
	TemplateKeyAMQPConnectorIdleTimeout string = "%s.idle_timeout"
	// TemplateKeyAMQPConnectorWriteTimeout Лимит времени на запись фрейма в брокер
	TemplateKeyAMQPConnectorWriteTimeout string = "%s.write_timeout"
	// TemplateKeyAMQPConnectorShutdownTimeout Лимит времени на мягкое освобождение ресурсов канала
	TemplateKeyAMQPConnectorShutdownTimeout string = "%s.shutdown_timeout"
)

// ====================================================================
// amqp connector
// ====================================================================
// Настройки по умолчанию для сетевого пула соединений брокеров сообщений по протоколу AMQP 0-9-1.
const (
	DefaultAMQPConnectorURL             string        = "amqp://localhost:5672/" // Сетевой адрес брокера AMQP v1.0 по умолчанию
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

// BuildDefaultAMQPConnectorSettings AMQP connector settings builder
//
// # ВНИМАНИЕ!
//
// Сюда не входят:
//
//	TemplateKeyAMQPConnectorUsername TemplateFlagAMQPConnectorUsername
//	TemplateKeyAMQPConnectorPassword TemplateFlagAMQPConnectorPassword
func BuildDefaultAMQPConnectorSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeyAMQPConnectorURL, keyPrefix), fmt.Sprintf(TemplateFlagAMQPConnectorURL, flagPrefix), DefaultAMQPConnectorURL, fmt.Sprintf("%s AMQP server URL", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyAMQPConnectorUsername, keyPrefix), fmt.Sprintf(TemplateFlagAMQPConnectorUsername, flagPrefix), "", fmt.Sprintf("%s AMQP server username", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyAMQPConnectorPassword, keyPrefix), fmt.Sprintf(TemplateFlagAMQPConnectorPassword, flagPrefix), "", fmt.Sprintf("%s AMQP server password", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAMQPConnectorConnectTimeout, keyPrefix), fmt.Sprintf(TemplateFlagAMQPConnectorConnectTimeout, flagPrefix), DefaultAMQPReceiverConnectTimeout, fmt.Sprintf("%s AMQP connect timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAMQPConnectorIdleTimeout, keyPrefix), fmt.Sprintf(TemplateFlagAMQPConnectorIdleTimeout, flagPrefix), DefaultAMQPConnectorIdleTimeout, fmt.Sprintf("%s AMQP idle timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAMQPConnectorWriteTimeout, keyPrefix), fmt.Sprintf(TemplateFlagAMQPConnectorWriteTimeout, flagPrefix), DefaultAMQPConnectorWriteTimeout, fmt.Sprintf("%s AMQP write timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAMQPConnectorShutdownTimeout, keyPrefix), fmt.Sprintf(TemplateFlagAMQPConnectorShutdownTimeout, flagPrefix), DefaultAMQPConnectorShutdownTimeout, fmt.Sprintf("%s AMQP shutdown timeout", flagPrefix)),
	}
}
