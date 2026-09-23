package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// KafkaSenderConfig содержит настройки для отправки (публикации) сообщений в брокер Kafka.
type KafkaSenderConfig struct {
	// Brokers хранит срез хостов брокеров кластера Kafka (например, ["kafka-node1:9092", "kafka-node2:9092"]).
	Brokers []string `mapstructure:"brokers" json:"brokers,omitempty" yaml:"brokers,omitempty"`

	// TargetName определяет имя целевого топика (Topic) в Kafka, куда будут отправляться сообщения.
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"`

	// ConnectTimeout задает ограничение по времени на установку сетевого соединения с брокерами.
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`

	// ShutdownTimeout определяет время, выделяемое врайтеру на плавное закрытие (включая сброс буферов на диски брокеров).
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"`

	// PublishMaxTryAttempts — максимальное количество попыток публикации сообщения при сетевых сбоях (Network Flaps).
	PublishMaxTryAttempts int `mapstructure:"publish_max_try_attempts" json:"publish_max_try_attempts" yaml:"publish_max_try_attempts"`

	// PublishBaseRetryDelay — начальная задержка перед первой повторной отправкой (используется для экспоненциального бэкоффа).
	PublishBaseRetryDelay time.Duration `mapstructure:"publish_base_retry_delay" json:"publish_base_retry_delay,omitempty" yaml:"publish_base_retry_delay,omitempty"`

	// PublishMaxRetryDelay — жесткий верхний лимит задержки между повторными попытками отправки.
	PublishMaxRetryDelay time.Duration `mapstructure:"publish_max_retry_delay" json:"publish_max_retry_delay" yaml:"publish_max_retry_delay"`

	// Безопасность и Аутентификация (SASL/PLAIN + TLS)
	Username string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`

	// Тонкие настройки асинхронного пакетирования (Батчинга) рантайма kafka-go
	BatchSize    int           `mapstructure:"batch_size" json:"batch_size,omitempty" yaml:"batch_size,omitempty"`
	BatchBytes   int           `mapstructure:"batch_bytes" json:"batch_bytes,omitempty" yaml:"batch_bytes,omitempty"`
	BatchTimeout time.Duration `mapstructure:"batch_timeout" json:"batch_timeout,omitempty" yaml:"batch_timeout,omitempty"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout,omitempty" yaml:"write_timeout,omitempty"`
	RequiredAcks int           `mapstructure:"required_acks" json:"required_acks,omitempty" yaml:"required_acks,omitempty"`
}

// NewKafkaSenderConfig — полный конструктор структуры конфигурации отправителя.
func NewKafkaSenderConfig(
	brokers []string,
	targetName string,
	connectTimeout time.Duration,
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

// NewDefaultKafkaSenderConfig конструктор структуры конфигурации со значениями по умолчанию
func NewDefaultKafkaSenderConfig() *KafkaSenderConfig {
	return NewKafkaSenderConfig(
		DefaultKafkaBrokers,
		"",
		DefaultKafkaSenderConnectTimeout,
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

// Validate выполняет строгую проверку входящих параметров конфигурации отправителя.
func (ksc *KafkaSenderConfig) Validate() error {
	if len(ksc.Brokers) == 0 {
		return errs.NewConfigValidateError("kafka sender", "Brokers", "at least one broker address is required", nil)
	}
	if strings.TrimSpace(ksc.TargetName) == "" {
		return errs.NewConfigValidateError("kafka sender", "TargetName", "empty", nil)
	}
	if !(ksc.ConnectTimeout > 0) {
		return errs.NewConfigValidateError("kafka sender", "ConnectTimeout", "less than 0", nil)
	}
	if !(ksc.ShutdownTimeout > 0) {
		return errs.NewConfigValidateError("kafka sender", "ShutdownTimeout", "less than 0", nil)
	}
	if !(ksc.PublishMaxTryAttempts >= 1) {
		return errs.NewConfigValidateError("kafka sender", "PublishMaxTryAttempts", "less than 1", nil)
	}
	if ksc.PublishBaseRetryDelay > ksc.PublishMaxRetryDelay {
		return errs.NewConfigValidateError("kafka sender", "PublishMaxRetryDelay", "less than base delay", nil)
	}
	if ksc.BatchSize <= 0 || ksc.BatchBytes <= 0 {
		return errs.NewConfigValidateError("kafka sender", "BatchSize/BatchBytes", "must be greater than 0", nil)
	}
	// Валидируем диапазон acks (разрешены только -1, 0, 1)
	if ksc.RequiredAcks < -1 || ksc.RequiredAcks > 1 {
		return errs.NewConfigValidateError("kafka sender", "RequiredAcks", "invalid value (must be -1, 0 or 1)", nil)
	}

	return nil
}
