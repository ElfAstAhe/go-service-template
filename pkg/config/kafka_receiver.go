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

	// Partition - идентификация партиции потребителя (Direct Consumer)
	Partition int `mapstructure:"partition" json:"partition,omitempty" yaml:"partition,omitempty"`

	// ConnectTimeout — таймаут на первичное подключение к координатору группы брокера.
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`

	// ShutdownTimeout — время на безопасную остановку чтения и корректный выход из Consumer Group.
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"`

	// MinBytes — минимальный объем данных (в байтах), который брокер должен собрать перед ответом клиенту.
	MinBytes int `mapstructure:"min_bytes" json:"min_bytes,omitempty" yaml:"min_bytes,omitempty"`

	// MaxBytes — максимальный объем данных, который клиент готов принять за одну сетевую трансляцию (итерацию Fetch).
	MaxBytes int `mapstructure:"max_bytes" json:"max_bytes,omitempty" yaml:"max_bytes,omitempty"`

	// MaxWait — максимальное время ожидания брокера, если объем накопленных сообщений еще не достиг лимита MinBytes.
	MaxWait time.Duration `mapstructure:"max_wait" json:"max_wait,omitempty" yaml:"max_wait,omitempty"`

	// Безопасность и Аутентификация (SASL/PLAIN + TLS)
	Username string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`

	// Расширенные таймауты координации группы и сокетов
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval" json:"heartbeat_interval,omitempty" yaml:"heartbeat_interval,omitempty"`
	SessionTimeout    time.Duration `mapstructure:"session_timeout" json:"session_timeout,omitempty" yaml:"session_timeout,omitempty"`
	RebalanceTimeout  time.Duration `mapstructure:"rebalance_timeout" json:"rebalance_timeout,omitempty" yaml:"rebalance_timeout,omitempty"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout" json:"read_timeout,omitempty" yaml:"read_timeout,omitempty"`

	// Тонкие настройки производительности рантайма kafka-go
	MaxAttempts   int    `mapstructure:"max_attempts" json:"max_attempts,omitempty" yaml:"max_attempts,omitempty"`
	QueueCapacity int    `mapstructure:"queue_capacity" json:"queue_capacity,omitempty" yaml:"queue_capacity,omitempty"`
	StartOffset   string `mapstructure:"start_offset" json:"start_offset,omitempty" yaml:"start_offset,omitempty"`
}

// NewKafkaReceiverConfig — полный конструктор структуры конфигурации получателя.
func NewKafkaReceiverConfig(
	brokers []string,
	targetName,
	groupID string,
	partition int,
	connectTimeout,
	shutdownTimeout time.Duration,
	minBytes int,
	maxBytes int,
	maxWait time.Duration,
	username string,
	password string,
	heartbeatInterval time.Duration,
	sessionTimeout time.Duration,
	rebalanceTimeout time.Duration,
	readTimeout time.Duration,
	maxAttempts int,
	queueCapacity int,
	startOffset string,
) *KafkaReceiverConfig {
	return &KafkaReceiverConfig{
		Brokers:           brokers,
		TargetName:        targetName,
		GroupID:           groupID,
		Partition:         partition,
		ConnectTimeout:    connectTimeout,
		ShutdownTimeout:   shutdownTimeout,
		MinBytes:          minBytes,
		MaxBytes:          maxBytes,
		MaxWait:           maxWait,
		Username:          username,
		Password:          password,
		HeartbeatInterval: heartbeatInterval,
		SessionTimeout:    sessionTimeout,
		RebalanceTimeout:  rebalanceTimeout,
		ReadTimeout:       readTimeout,
		MaxAttempts:       maxAttempts,
		QueueCapacity:     queueCapacity,
		StartOffset:       startOffset,
	}
}

func NewDefaultKafkaReceiverConfig() *KafkaReceiverConfig {
	return NewKafkaReceiverConfig(
		DefaultKafkaBrokers,
		"",
		"",
		-1,
		DefaultKafkaReceiverConnectTimeout,
		DefaultKafkaReceiverShutdownTimeout,
		DefaultKafkaReceiverMinBytes,
		DefaultKafkaReceiverMaxBytes,
		DefaultKafkaReceiverMaxWait,
		"",
		"",
		DefaultKafkaReceiverHeartbeatInterval,
		DefaultKafkaReceiverSessionTimeout,
		DefaultKafkaReceiverRebalanceTimeout,
		DefaultKafkaReceiverReadTimeout,
		DefaultKafkaReceiverMaxAttempts,
		DefaultKafkaReceiverQueueCapacity,
		DefaultKafkaReceiverStartOffset,
	)
}

// Validate выполняет строгую валидацию входящих параметров конфигурации получателя.
func (krc *KafkaReceiverConfig) Validate() error {
	if len(krc.Brokers) == 0 {
		return errs.NewConfigValidateError("kafka receiver", "Brokers", "at least one broker address is required", nil)
	}
	if strings.TrimSpace(krc.TargetName) == "" {
		return errs.NewConfigValidateError("kafka receiver", "TargetName", "empty (required parameter for start)", nil)
	}

	hasGroup := strings.TrimSpace(krc.GroupID) != ""
	hasPartition := krc.Partition >= 0 // ИСПРАВЛЕНО: Любое значение >= 0 означает, что партиция явно указана

	// Проверка XOR: должно быть заполнено ровно одно из двух
	if !hasGroup && !hasPartition {
		return errs.NewConfigValidateError("kafka receiver", "GroupID/Partition", "either GroupID must be set or Partition must be >= 0", nil)
	}
	if hasGroup && hasPartition {
		return errs.NewConfigValidateError("kafka receiver", "GroupID/Partition", "GroupID and Partition are mutually exclusive options", nil)
	}
	if !(krc.ConnectTimeout > 0) {
		return errs.NewConfigValidateError("kafka receiver", "ConnectTimeout", "less than or equal to 0", nil)
	}
	if !(krc.ShutdownTimeout > 0) {
		return errs.NewConfigValidateError("kafka receiver", "ShutdownTimeout", "less than or equal to 0", nil)
	}
	if krc.MinBytes <= 0 || krc.MaxBytes <= 0 {
		return errs.NewConfigValidateError("kafka receiver", "MinBytes/MaxBytes", "must be greater than 0", nil)
	}
	// Золотое правило Kafka (только для режима группы): сессионный таймаут должен вмещать минимум 3 попытки отправки heartbeat пингов
	if hasGroup && krc.SessionTimeout < krc.HeartbeatInterval*3 {
		return errs.NewConfigValidateError("kafka receiver", "SessionTimeout", "must be at least 3 times greater than heartbeat interval", nil)
	}
	if krc.ReadTimeout <= krc.MaxWait {
		return errs.NewConfigValidateError("kafka receiver", "ReadTimeout", "must be strictly greater than MaxWait to prevent socket EOF on low-volume topics", nil)
	}
	startOffsetLower := strings.ToLower(strings.TrimSpace(krc.StartOffset))
	if startOffsetLower != "first" && startOffsetLower != "last" {
		return errs.NewConfigValidateError("kafka receiver", "StartOffset", "must be strictly 'first' or 'last'", nil)
	}

	return nil
}
