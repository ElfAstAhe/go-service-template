package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// flags templates
const (
	TemplateFlagKafkaSenderBrokers               string = "%s-brokers"
	TemplateFlagKafkaSenderTargetName            string = "%s-target-name"
	TemplateFlagKafkaSenderConnectTimeout        string = "%s-connect-timeout"
	TemplateFlagKafkaSenderIdleTimeout           string = "%s-idle-timeout"
	TemplateFlagKafkaSenderShutdownTimeout       string = "%s-shutdown-timeout"
	TemplateFlagKafkaSenderPublishMaxTryAttempts string = "%s-publish-max-try-attempts"
	TemplateFlagKafkaSenderPublishBaseRetryDelay string = "%s-publish-base-retry-delay"
	TemplateFlagKafkaSenderPublishMaxRetryDelay  string = "%s-publish-max-retry-delay"
	TemplateFlagKafkaSenderUsername              string = "%s-username"
	TemplateFlagKafkaSenderPassword              string = "%s-password"
	TemplateFlagKafkaSenderBatchSize             string = "%s-batch-size"
	TemplateFlagKafkaSenderBatchBytes            string = "%s-batch-bytes"
	TemplateFlagKafkaSenderBatchTimeout          string = "%s-batch-timeout"
	TemplateFlagKafkaSenderWriteTimeout          string = "%s-write-timeout"
	TemplateFlagKafkaSenderRequiredAcks          string = "%s-required-acks"
)

// keys templates
const (
	TemplateKeyKafkaSenderBrokers               string = "%s.brokers"
	TemplateKeyKafkaSenderTargetName            string = "%s.target_name"
	TemplateKeyKafkaSenderConnectTimeout        string = "%s.connect_timeout"
	TemplateKeyKafkaSenderIdleTimeout           string = "%s.idle_timeout"
	TemplateKeyKafkaSenderShutdownTimeout       string = "%s.shutdown_timeout"
	TemplateKeyKafkaSenderPublishMaxTryAttempts string = "%s.publish_max_try_attempts"
	TemplateKeyKafkaSenderPublishBaseRetryDelay string = "%s.publish_base_retry_delay"
	TemplateKeyKafkaSenderPublishMaxRetryDelay  string = "%s.publish_max_retry_delay"
	TemplateKeyKafkaSenderUsername              string = "%s.username"
	TemplateKeyKafkaSenderPassword              string = "%s.password"
	TemplateKeyKafkaSenderBatchSize             string = "%s.batch_size"
	TemplateKeyKafkaSenderBatchBytes            string = "%s.batch_bytes"
	TemplateKeyKafkaSenderBatchTimeout          string = "%s.batch_timeout"
	TemplateKeyKafkaSenderWriteTimeout          string = "%s.write_timeout"
	TemplateKeyKafkaSenderRequiredAcks          string = "%s.required_acks"
)

// ====================================================================
// Kafka Sender (Producer) Defaults
// ====================================================================
// Системные параметры по умолчанию для высокопроизводительного асинхронного продюсера Кафки.
const (
	// DefaultKafkaSenderConnectTimeout задает лимит времени на установку сетевого соединения с брокерами.
	DefaultKafkaSenderConnectTimeout time.Duration = 10 * time.Second

	// DefaultKafkaSenderIdleTimeout задаёт время простоя сетевых соединений в пуле коннектов.
	DefaultKafkaSenderIdleTimeout time.Duration = 60 * time.Second

	// DefaultKafkaSenderShutdownTimeout определяет время, выделяемое врайтеру на плавное закрытие.
	// 15 секунд гарантируют успешный сброс (flushing) асинхронных буферов из памяти на диски брокеров при остановке пода.
	DefaultKafkaSenderShutdownTimeout time.Duration = 15 * time.Second

	// DefaultKafkaSenderPublishMaxTryAttempts — количество попыток публикации сообщения при сетевых сбоях (Network Flaps).
	// Повторами на транспортном уровне мы управляем сами в методе Publish.
	DefaultKafkaSenderPublishMaxTryAttempts int = 2

	// DefaultKafkaSenderPublishBaseRetryDelay — начальная задержка перед первой повторной отправкой для экспоненциального бэкоффа.
	DefaultKafkaSenderPublishBaseRetryDelay time.Duration = 100 * time.Millisecond

	// DefaultKafkaSenderPublishMaxRetryDelay — жесткий верхний лимит задержки между повторными попытками отправки.
	DefaultKafkaSenderPublishMaxRetryDelay time.Duration = 3 * time.Second

	// ПОЛЯ ПРОИЗВОДИТЕЛЬНОСТИ (Асинхронный батчинг)

	// DefaultKafkaSenderBatchSize — лимит количества сообщений в локальном буфере перед отправкой пачки.
	DefaultKafkaSenderBatchSize int = 100

	// DefaultKafkaSenderBatchBytes — максимальный объем одной пачки в байтах (1 MB).
	DefaultKafkaSenderBatchBytes int = 1048576

	// DefaultKafkaSenderBatchTimeout — 10ms. Время ожидания накопления пачки.
	// Исключает секундную задержку библиотеки по умолчанию, обеспечивая минимальный latency.
	DefaultKafkaSenderBatchTimeout time.Duration = 10 * time.Millisecond

	// DefaultKafkaSenderWriteTimeout — жесткий таймаут сокета на операцию отправки пачки в сеть.
	DefaultKafkaSenderWriteTimeout time.Duration = 10 * time.Second

	// DefaultKafkaSenderRequiredAcks — уровень подтверждения записи брокером (-1 означает "all" — весь ISR пул).
	DefaultKafkaSenderRequiredAcks int = -1
)

// KafkaSenderConfig содержит настройки для отправки (публикации) сообщений в брокер Kafka (Producer Config).
//
// Инкапсулирует параметры сетевого пула соединений, тонкие настройки асинхронного батчинга,
// политики экспоненциального бэккоффа повторных попыток отправки, SASL/TLS секреты и уровни квитирования (Acks).
type KafkaSenderConfig struct {
	// Brokers хранит срез хостов брокеров кластера Kafka (например, ["kafka-node1:9092", "kafka-node2:9092"]).
	Brokers []string `mapstructure:"brokers" json:"brokers,omitempty" yaml:"brokers,omitempty" toml:"brokers,omitempty"`

	// TargetName определяет имя целевого топика (Topic) в Kafka, куда будут отправляться сообщения.
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty" toml:"target_name,omitempty"`

	// ConnectTimeout задает ограничение по времени на установку сетевого соединения с брокерами.
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty" toml:"connect_timeout,omitempty"`

	// IdleTimeout задаёт максимальное время простоя открытого соединения в пуле коннектов.
	IdleTimeout time.Duration `mapstructure:"idle_timeout" json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty" toml:"idle_timeout,omitempty"`

	// ShutdownTimeout определяет время, выделяемое врайтеру на плавное закрытие (включая сброс буферов на диски брокеров).
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty" toml:"shutdown_timeout,omitempty"`

	// PublishMaxTryAttempts — максимальное количество попыток публикации сообщения при сетевых сбоях (Network Flaps).
	PublishMaxTryAttempts int `mapstructure:"publish_max_try_attempts" json:"publish_max_try_attempts" yaml:"publish_max_try_attempts" toml:"publish_max_try_attempts,omitempty"`

	// PublishBaseRetryDelay — начальная задержка перед первой повторной отправкой (используется для экспоненциального бэкоффа).
	PublishBaseRetryDelay time.Duration `mapstructure:"publish_base_retry_delay" json:"publish_base_retry_delay,omitempty" yaml:"publish_base_retry_delay,omitempty" toml:"publish_base_retry_delay,omitempty"`

	// PublishMaxRetryDelay — жесткий верхний лимит задержки между повторными попытками отправки.
	PublishMaxRetryDelay time.Duration `mapstructure:"publish_max_retry_delay" json:"publish_max_retry_delay" yaml:"publish_max_retry_delay" toml:"publish_max_retry_delay,omitempty"`

	// Безопасность и Аутентификация (SASL/PLAIN + TLS контур)
	Username string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty" toml:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`

	// Тонкие настройки асинхронного пакетирования (Батчинга) рантайма библиотеки kafka-go

	// BatchSize Лимит количества сообщений в локальном буфере перед отправкой пачки
	BatchSize int `mapstructure:"batch_size" json:"batch_size,omitempty" yaml:"batch_size,omitempty" toml:"batch_size,omitempty"`
	// BatchBytes Максимальный объем одной пачки в байтах перед сбросом в сеть
	BatchBytes int `mapstructure:"batch_bytes" json:"batch_bytes,omitempty" yaml:"batch_bytes,omitempty" toml:"batch_bytes,omitempty"`
	// BatchTimeout Время ожидания накопления пачки (предотвращает зависание при низком RPS)
	BatchTimeout time.Duration `mapstructure:"batch_timeout" json:"batch_timeout,omitempty" yaml:"batch_timeout,omitempty" toml:"batch_timeout,omitempty"`
	// WriteTimeout Жесткий таймаут сокета на непосредственную операцию записи пачки в сеть
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty" toml:"write_timeout,omitempty"`
	// RequiredAcks Уровень подтверждения записи брокером (-1 = all, 0 = none, 1 = leader)
	RequiredAcks int `mapstructure:"required_acks" json:"required_acks,omitempty" yaml:"required_acks,omitempty" toml:"required_acks,omitempty"`
}

// NewKafkaSenderConfig — полный конструктор структуры конфигурации отправителя.
func NewKafkaSenderConfig(
	brokers []string,
	targetName string,
	connectTimeout time.Duration,
	idleTimeout time.Duration,
	shutdownTimeout time.Duration,
	publishMaxTryAttempts int,
	publishBaseRetryDelay time.Duration,
	publishMaxRetryDelay time.Duration,
	username string,
	password string,
	batchSize int,
	batchBytes int,
	batchTimeout time.Duration,
	writeTimeout time.Duration,
	requiredAcks int,
) *KafkaSenderConfig {
	return &KafkaSenderConfig{
		Brokers:               brokers,
		TargetName:            targetName,
		ConnectTimeout:        connectTimeout,
		IdleTimeout:           idleTimeout,
		ShutdownTimeout:       shutdownTimeout,
		PublishMaxTryAttempts: publishMaxTryAttempts,
		PublishBaseRetryDelay: publishBaseRetryDelay,
		PublishMaxRetryDelay:  publishMaxRetryDelay,
		Username:              username,
		Password:              password,
		BatchSize:             batchSize,
		BatchBytes:            batchBytes,
		BatchTimeout:          batchTimeout,
		WriteTimeout:          writeTimeout,
		RequiredAcks:          requiredAcks,
	}
}

// NewDefaultKafkaSenderConfig конструктор структуры конфигурации со значениями по умолчанию фреймворка.
func NewDefaultKafkaSenderConfig() *KafkaSenderConfig {
	return NewKafkaSenderConfig(
		DefaultKafkaBrokers,
		"", // Намеренно оставляем пустым для обязательного переопределения через файлы конфигурации среды
		DefaultKafkaSenderConnectTimeout,
		DefaultKafkaSenderIdleTimeout,
		DefaultKafkaSenderShutdownTimeout,
		DefaultKafkaSenderPublishMaxTryAttempts,
		DefaultKafkaSenderPublishBaseRetryDelay,
		DefaultKafkaSenderPublishMaxRetryDelay,
		"",
		"",
		DefaultKafkaSenderBatchSize,
		DefaultKafkaSenderBatchBytes,
		DefaultKafkaSenderBatchTimeout,
		DefaultKafkaSenderWriteTimeout,
		DefaultKafkaSenderRequiredAcks,
	)
}

// Validate выполняет строгую семантическую и математическую проверку входящих параметров конфигурации отправителя.
// Полностью пресекает запуск продюсера кластера Kafka с невалидными таймаутами, нарушенными ретраями или некорректным ack-квитированием.
func (ksc *KafkaSenderConfig) Validate() error {
	if len(ksc.Brokers) == 0 {
		return errs.NewConfigValidateError("kafka sender", "Brokers", "at least one broker address is required", nil)
	}
	if strings.TrimSpace(ksc.TargetName) == "" {
		return errs.NewConfigValidateError("kafka sender", "TargetName", "empty", nil)
	}
	if ksc.ConnectTimeout <= 0 {
		return errs.NewConfigValidateError("kafka sender", "ConnectTimeout", "less than 0", nil)
	}
	if ksc.IdleTimeout <= 0 {
		return errs.NewConfigValidateError("kafka sender", "IdleTimeout", "less than 0", nil)
	}
	if ksc.ShutdownTimeout <= 0 {
		return errs.NewConfigValidateError("kafka sender", "ShutdownTimeout", "less than 0", nil)
	}
	if ksc.PublishMaxTryAttempts < 1 {
		return errs.NewConfigValidateError("kafka sender", "PublishMaxTryAttempts", "less than 1", nil)
	}
	if ksc.PublishBaseRetryDelay > ksc.PublishMaxRetryDelay {
		return errs.NewConfigValidateError("kafka sender", "PublishMaxRetryDelay", "less than base delay", nil)
	}
	if ksc.BatchSize <= 0 || ksc.BatchBytes <= 0 {
		return errs.NewConfigValidateError("kafka sender", "BatchSize/BatchBytes", "must be greater than 0", nil)
	}
	// Валидируем жестко регламентированный спецификацией Apache Kafka диапазон acks (разрешены только -1, 0, 1)
	if ksc.RequiredAcks < -1 || ksc.RequiredAcks > 1 {
		return errs.NewConfigValidateError("kafka sender", "RequiredAcks", "invalid value (must be -1, 0 or 1)", nil)
	}

	return nil
}

// BuildDefaultKafkaSenderSettings kafka sender settings builder
//
// # ВНИМАНИЕ!!!
//
// Сюда не входят:
//
// TemplateKeyKafkaSenderTargetName TemplateFlagKafkaSenderTargetName
// TemplateKeyKafkaSenderUsername TemplateFlagKafkaSenderUsername
// TemplateKeyKafkaSenderPassword TemplateFlagKafkaSenderPassword
func BuildDefaultKafkaSenderSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderBrokers, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderBrokers, flagPrefix), DefaultKafkaBrokers, fmt.Sprintf("%s kafka brokers, separated by comma, like: localhost:9092,localhost:9093", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderTargetName, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderTargetName, flagPrefix), "", fmt.Sprintf("%s kafka queue/topic name", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderConnectTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderConnectTimeout, flagPrefix), DefaultKafkaSenderConnectTimeout, fmt.Sprintf("%s kafka connect timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderIdleTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderIdleTimeout, flagPrefix), DefaultKafkaSenderIdleTimeout, fmt.Sprintf("%s kafka idle timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderShutdownTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderShutdownTimeout, flagPrefix), DefaultKafkaSenderShutdownTimeout, fmt.Sprintf("%s kafka shutdown timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderPublishMaxTryAttempts, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderPublishMaxTryAttempts, flagPrefix), DefaultKafkaSenderPublishMaxTryAttempts, fmt.Sprintf("%s kafka publish max try attempts", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderPublishBaseRetryDelay, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderPublishBaseRetryDelay, flagPrefix), DefaultKafkaSenderPublishMaxTryAttempts, fmt.Sprintf("%s kafka publish base retry delay", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderPublishMaxRetryDelay, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderPublishMaxRetryDelay, flagPrefix), DefaultKafkaSenderPublishMaxRetryDelay, fmt.Sprintf("%s kafka publish max retry delay", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderUsername, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderUsername, flagPrefix), "", fmt.Sprintf("%s kafka username", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderPassword, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderPassword, flagPrefix), "", fmt.Sprintf("%s kafka password", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderBatchSize, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderBatchSize, flagPrefix), DefaultKafkaSenderBatchSize, fmt.Sprintf("%s kafka batch size", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderBatchBytes, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderBatchBytes, flagPrefix), DefaultKafkaSenderBatchBytes, fmt.Sprintf("%s kafka batch bytes", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderBatchTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderBatchTimeout, flagPrefix), DefaultKafkaSenderBatchTimeout, fmt.Sprintf("%s kafka batch timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderWriteTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderWriteTimeout, flagPrefix), DefaultKafkaSenderWriteTimeout, fmt.Sprintf("%s kafka write timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaSenderRequiredAcks, keyPrefix), fmt.Sprintf(TemplateFlagKafkaSenderRequiredAcks, flagPrefix), DefaultKafkaSenderRequiredAcks, fmt.Sprintf("%s kafka required acks", flagPrefix)),
	}
}
