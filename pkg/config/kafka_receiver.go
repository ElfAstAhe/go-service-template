package config

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// KafkaReceiverConfig содержит настройки для вычитки сообщений и координации в рамках Consumer Group.
type KafkaReceiverConfig struct {
	// Brokers хранит срез хостов кластера Kafka для инициализации коннекта.
	Brokers []string `mapstructure:"brokers" json:"brokers,omitempty" yaml:"brokers,omitempty"`

	// TargetName определяет топик брокера, на который подписывается данный слушатель.
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty"`

	// GroupID — идентификатор группы потребителей (Consumer Group). Ключевой параметр Kafka для балансировки партиций.
	GroupID string `mapstructure:"group_id" json:"group_id,omitempty" yaml:"group_id,omitempty"`

	// ConnectTimeout — таймаут на первичное подключение к координатору группы брокера.
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`

	// ShutdownTimeout — время на безопасную остановку чтения и корректный выход из Consumer Group.
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"`

	// MinBytes — минимальный объем данных (в байтах), который брокер должен собрать перед ответом клиенту.
	// Используется для оптимизации сетевого трафика (пакетирование логов на стороне брокера).
	MinBytes int `mapstructure:"min_bytes" json:"min_bytes,omitempty" yaml:"min_bytes,omitempty"`

	// MaxBytes — максимальный объем данных, который клиент готов принять за одну сетевую трансляцию (итерацию Fetch).
	MaxBytes int `mapstructure:"max_bytes" json:"max_bytes,omitempty" yaml:"max_bytes,omitempty"`

	// MaxWait — максимальное время ожидания брокера, если объем накопленных сообщений еще не достиг лимита MinBytes.
	// Предотвращает бесконечную блокировку горутины чтения при слабом потоке сообщений.
	MaxWait time.Duration `mapstructure:"max_wait" json:"max_wait,omitempty" yaml:"max_wait,omitempty"`

	// Безопасность и Аутентификация (SASL/PLAIN + TLS)
	Username           string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password           string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`
	InsecureConnection bool   `mapstructure:"insecure_connection" json:"insecure_connection,omitempty" yaml:"insecure_connection,omitempty"`
}

func NewKafkaReceiverConfig(
	brokers []string,
	targetName string,
	groupID string,
	connectTimeout time.Duration,
	shutdownTimeout time.Duration,
	minBytes int,
	maxBytes int,
	maxWait time.Duration,
	username string,
	password string,
	insecureConnection bool,
) *KafkaReceiverConfig {
	return &KafkaReceiverConfig{
		Brokers:            brokers,
		TargetName:         targetName,
		GroupID:            groupID,
		ConnectTimeout:     connectTimeout,
		ShutdownTimeout:    shutdownTimeout,
		MinBytes:           minBytes,
		MaxBytes:           maxBytes,
		MaxWait:            maxWait,
		Username:           username,
		Password:           password,
		InsecureConnection: insecureConnection,
	}
}

func NewDefaultKafkaReceiverConfig() *KafkaReceiverConfig {
	return NewKafkaReceiverConfig(
		DefaultKafkaBrokers,
		"",
		"",
		DefaultKafkaReceiverConnectTimeout,
		DefaultKafkaReceiverShutdownTimeout,
		DefaultKafkaReceiverMinBytes,
		DefaultKafkaReceiverMaxBytes,
		DefaultKafkaReceiverMaxWait,
		"",
		"",
		DefaultKafkaReceiverInsecureConnection,
	)
}

// Validate выполняет строгую проверку входящих параметров конфигурации получателя.
func (krc *KafkaReceiverConfig) Validate() error {
	if len(krc.Brokers) == 0 {
		return errs.NewConfigValidateError("kafka receiver", "Brokers", "at least one broker address is required", nil)
	}
	if strings.TrimSpace(krc.TargetName) == "" {
		return errs.NewConfigValidateError("kafka receiver", "TargetName", "empty", nil)
	}
	if strings.TrimSpace(krc.GroupID) == "" {
		return errs.NewConfigValidateError("kafka receiver", "GroupID", "empty", nil)
	}
	if !(krc.ConnectTimeout > 0) {
		return errs.NewConfigValidateError("kafka receiver", "ConnectTimeout", "less than 0", nil)
	}
	if !(krc.ShutdownTimeout > 0) {
		return errs.NewConfigValidateError("kafka receiver", "ShutdownTimeout", "less than 0", nil)
	}

	return nil
}
