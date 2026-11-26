package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	// TemplateFlagKafkaReceiverBrokers брокеры
	TemplateFlagKafkaReceiverBrokers string = "%s-brokers"
	// TemplateFlagKafkaReceiverTargetName topic/queue
	TemplateFlagKafkaReceiverTargetName string = "%s-target-name"
	// TemplateFlagKafkaReceiverGroupID идентификатор группы
	TemplateFlagKafkaReceiverGroupID string = "%s-group-id"
	// TemplateFlagKafkaReceiverPartition номер партиции
	TemplateFlagKafkaReceiverPartition string = "%s-partition"
	// TemplateFlagKafkaReceiverConnectTimeout задает лимит времени на подключение к координатору группы брокеров
	TemplateFlagKafkaReceiverConnectTimeout string = "%s-connect-timeout"
	// TemplateFlagKafkaReceiverShutdownTimeout время на безопасную остановку чтения и корректный выход из Consumer Group
	TemplateFlagKafkaReceiverShutdownTimeout string = "%s-shutdown-timeout"
	// TemplateFlagKafkaReceiverMinBytes минимальный объем данных (в байтах), который брокер должен собрать перед ответом
	TemplateFlagKafkaReceiverMinBytes string = "%s-min-bytes"
	// TemplateFlagKafkaReceiverMaxBytes максимальный объем данных, принимаемый за одну сетевую трансляцию (итерацию Fetch)
	TemplateFlagKafkaReceiverMaxBytes string = "%s-max-bytes"
	// TemplateFlagKafkaReceiverMaxWait максимальное время ожидания брокера, если объем данных еще не достиг лимита MinBytes
	TemplateFlagKafkaReceiverMaxWait string = "%s-max-wait"
	// TemplateFlagKafkaReceiverUsername пользователь
	TemplateFlagKafkaReceiverUsername string = "%s-username"
	// TemplateFlagKafkaReceiverPassword пароль
	TemplateFlagKafkaReceiverPassword string = "%s-password"
	// TemplateFlagKafkaReceiverHeartbeatInterval частота фонового пинга ("я жив") к координатору группы
	TemplateFlagKafkaReceiverHeartbeatInterval string = "%s-heartbeat-interval"
	// TemplateFlagKafkaReceiverSessionTimeout таймаут отсутствия пингов, после которого брокер считает под мертвым и запускает ребаланс
	TemplateFlagKafkaReceiverSessionTimeout string = "%s-session-timeout"
	// TemplateFlagKafkaReceiverRebalanceTimeout время, выделяемое воркеру на доработку текущей пачки и сдачу партиций при ребалансе
	TemplateFlagKafkaReceiverRebalanceTimeout string = "%s-rebalance-timeout"
	// TemplateFlagKafkaReceiverReadTimeout низкоуровневый таймаут сетевого сокета на чтение. Уходит в Dialer
	TemplateFlagKafkaReceiverReadTimeout string = "%s-read-timeout"
	// TemplateFlagKafkaReceiverMaxAttempts количество попыток переподключения до возврата критической ошибки
	TemplateFlagKafkaReceiverMaxAttempts string = "%s-max-attempts"
	// TemplateFlagKafkaReceiverQueueCapacity емкость внутреннего фонового буфера предвыборки сообщений из сети
	TemplateFlagKafkaReceiverQueueCapacity string = "%s-queue-capacity"
	// TemplateFlagKafkaReceiverStartOffset определяет точку старта, если для Consumer Group еще нет сохраненного оффсета ("first" или "last")
	TemplateFlagKafkaReceiverStartOffset string = "%s-start-offset"
)

const (
	// TemplateKeyKafkaReceiverBrokers брокеры
	TemplateKeyKafkaReceiverBrokers string = "%s.brokers"
	// TemplateKeyKafkaReceiverTargetName topic/queue
	TemplateKeyKafkaReceiverTargetName string = "%s.target_name"
	// TemplateKeyKafkaReceiverGroupID идентификатор группы
	TemplateKeyKafkaReceiverGroupID string = "%s.group_id"
	// TemplateKeyKafkaReceiverPartition номер партиции
	TemplateKeyKafkaReceiverPartition string = "%s.partition"
	// TemplateKeyKafkaReceiverConnectTimeout задает лимит времени на подключение к координатору группы брокеров
	TemplateKeyKafkaReceiverConnectTimeout string = "%s.connect_timeout"
	// TemplateKeyKafkaReceiverShutdownTimeout время на безопасную остановку чтения и корректный выход из Consumer Group
	TemplateKeyKafkaReceiverShutdownTimeout string = "%s.shutdown_timeout"
	// TemplateKeyKafkaReceiverMinBytes минимальный объем данных (в байтах), который брокер должен собрать перед ответом
	TemplateKeyKafkaReceiverMinBytes string = "%s.min_bytes"
	// TemplateKeyKafkaReceiverMaxBytes максимальный объем данных, принимаемый за одну сетевую трансляцию (итерацию Fetch)
	TemplateKeyKafkaReceiverMaxBytes string = "%s.max_bytes"
	// TemplateKeyKafkaReceiverMaxWait максимальное время ожидания брокера, если объем данных еще не достиг лимита MinBytes
	TemplateKeyKafkaReceiverMaxWait string = "%s.max_wait"
	// TemplateKeyKafkaReceiverUsername пользователь
	TemplateKeyKafkaReceiverUsername string = "%s.username"
	// TemplateKeyKafkaReceiverPassword пароль
	TemplateKeyKafkaReceiverPassword string = "%s.password"
	// TemplateKeyKafkaReceiverHeartbeatInterval частота фонового пинга ("я жив") к координатору группы
	TemplateKeyKafkaReceiverHeartbeatInterval string = "%s.heartbeat_interval"
	// TemplateKeyKafkaReceiverSessionTimeout таймаут отсутствия пингов, после которого брокер считает под мертвым и запускает ребаланс
	TemplateKeyKafkaReceiverSessionTimeout string = "%s.session_timeout"
	// TemplateKeyKafkaReceiverRebalanceTimeout время, выделяемое воркеру на доработку текущей пачки и сдачу партиций при ребалансе
	TemplateKeyKafkaReceiverRebalanceTimeout string = "%s.rebalance_timeout"
	// TemplateKeyKafkaReceiverReadTimeout низкоуровневый таймаут сетевого сокета на чтение. Уходит в Dialer
	TemplateKeyKafkaReceiverReadTimeout string = "%s.read_timeout"
	// TemplateKeyKafkaReceiverMaxAttempts количество попыток переподключения до возврата критической ошибки
	TemplateKeyKafkaReceiverMaxAttempts string = "%s.max_attempts"
	// TemplateKeyKafkaReceiverQueueCapacity емкость внутреннего фонового буфера предвыборки сообщений из сети
	TemplateKeyKafkaReceiverQueueCapacity string = "%s.queue_capacity"
	// TemplateKeyKafkaReceiverStartOffset определяет точку старта, если для Consumer Group еще нет сохраненного оффсета ("first" или "last")
	TemplateKeyKafkaReceiverStartOffset string = "%s.start_offset"
)

// ====================================================================
// Kafka Receiver (Consumer) Defaults
// ====================================================================
// Системные параметры по умолчанию для распределенного отказоустойчивого консьюмера Кафки.
const (
	// DefaultKafkaReceiverPartition идентификация партиции потребителя (Direct Consumer режим без координатора).
	DefaultKafkaReceiverPartition int = -1

	// DefaultKafkaReceiverConnectTimeout задает лимит времени на подключение к координатору группы брокеров.
	DefaultKafkaReceiverConnectTimeout time.Duration = 10 * time.Second

	// DefaultKafkaReceiverShutdownTimeout — время на безопасную остановку чтения и корректный выход из Consumer Group.
	DefaultKafkaReceiverShutdownTimeout time.Duration = 15 * time.Second

	// DefaultKafkaReceiverMinBytes — минимальный объем данных (в байтах), который брокер должен собрать перед ответом.
	// Снижено до 1024 (1 KB) для dev/test окружений, чтобы одиночные мелкие сообщения доставлялись без задержек.
	DefaultKafkaReceiverMinBytes int = 1024

	// DefaultKafkaReceiverMaxBytes — максимальный объем данных, принимаемый за одну сетевую трансляцию (итерацию Fetch).
	// Значение 10e6 (10 MB) защищает consumer-группу от застревания при обработке «жирных» JSON-пакетов.
	DefaultKafkaReceiverMaxBytes int = 10e6

	// DefaultKafkaReceiverMaxWait — максимальное время ожидания брокера, если объем данных еще не достиг лимита MinBytes.
	// Значение 500ms на dev/test экономит ресурсы CPU сервера, предотвращая «горячий цикл» пустых запросов.
	DefaultKafkaReceiverMaxWait time.Duration = 500 * time.Millisecond

	// DefaultKafkaReceiverHeartbeatInterval — частота фонового пинга ("я жив") к координатору группы.
	DefaultKafkaReceiverHeartbeatInterval time.Duration = 3 * time.Second

	// DefaultKafkaReceiverSessionTimeout — таймаут отсутствия пингов, после которого брокер считает под мертвым и запускает ребаланс.
	DefaultKafkaReceiverSessionTimeout time.Duration = 30 * time.Second

	// DefaultKafkaReceiverRebalanceTimeout — время, выделяемое воркеру на доработку текущей пачки и сдачу партиций при ребалансе.
	DefaultKafkaReceiverRebalanceTimeout time.Duration = 60 * time.Second

	// DefaultKafkaReceiverReadTimeout — низкоуровневый таймаут сетевого сокета на чтение. Уходит в Dialer.
	DefaultKafkaReceiverReadTimeout time.Duration = 10 * time.Second

	// DefaultKafkaReceiverMaxAttempts — количество попыток переподключения до возврата критической ошибки.
	DefaultKafkaReceiverMaxAttempts int = 3

	// DefaultKafkaReceiverQueueCapacity — емкость внутреннего фонового буфера предвыборки сообщений из сети.
	DefaultKafkaReceiverQueueCapacity int = 100

	// DefaultKafkaReceiverStartOffset определяет точку старта, если для Consumer Group еще нет сохраненного оффсета ("first" или "last").
	DefaultKafkaReceiverStartOffset string = "first"
)

// KafkaReceiverConfig содержит настройки для вычитки сообщений и координации в рамках Consumer Group кластера Kafka.
//
// Управляет сетевыми лимитами рантайма (таймауты, объемы батчей), стратегиями координации распределенной группы,
// параметрами безопасности SASL/TLS и тонкими настройками емкости внутреннего I/O буфера.
type KafkaReceiverConfig struct {
	// Brokers хранит срез хостов кластера Kafka для инициализации первичного коннекта.
	Brokers []string `mapstructure:"brokers" json:"brokers,omitempty" yaml:"brokers,omitempty" toml:"brokers,omitempty"`

	// TargetName определяет имя топика брокера, на который подписывается данный слушатель.
	TargetName string `mapstructure:"target_name" json:"target_name,omitempty" yaml:"target_name,omitempty" toml:"target_name,omitempty"`

	// GroupID — идентификатор группы потребителей (Consumer Group). Ключевой параметр Kafka для балансировки партиций.
	GroupID string `mapstructure:"group_id" json:"group_id,omitempty" yaml:"group_id,omitempty" toml:"group_id,omitempty"`

	// Partition - идентификация партиции потребителя (Direct Consumer режим без координатора)
	Partition int `mapstructure:"partition" json:"partition,omitempty" yaml:"partition,omitempty" toml:"partition,omitempty"`

	// ConnectTimeout — таймаут на первичное подключение к координатору группы брокера.
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty" toml:"connect_timeout,omitempty"`

	// ShutdownTimeout — время на безопасную остановку чтения и корректный выход из Consumer Group.
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty" toml:"shutdown_timeout,omitempty"`

	// MinBytes — минимальный объем данных (в байтах), который брокер должен собрать перед ответом клиенту.
	MinBytes int `mapstructure:"min_bytes" json:"min_bytes,omitempty" yaml:"min_bytes,omitempty" toml:"min_bytes,omitempty"`

	// MaxBytes — максимальный объем данных, который клиент готов принять за одну сетечную трансляцию (итерацию Fetch).
	MaxBytes int `mapstructure:"max_bytes" json:"max_bytes,omitempty" yaml:"max_bytes,omitempty" toml:"max_bytes,omitempty"`

	// MaxWait — максимальное время ожидания брокера, если объем накопленных сообщений еще не достиг лимита MinBytes.
	MaxWait time.Duration `mapstructure:"max_wait" json:"max_wait,omitempty" yaml:"max_wait,omitempty" toml:"max_wait,omitempty"`

	// Безопасность и Аутентификация (SASL/PLAIN + TLS контур)
	Username string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty" toml:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`

	// Расширенные таймауты координации группы и сокетов

	// HeartbeatInterval Частота фонового пинга ("я жив") к координатору группы
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval" json:"heartbeat_interval,omitempty" yaml:"heartbeat_interval,omitempty" toml:"heartbeat_interval,omitempty"`
	// SessionTimeout Таймаут отсутствия пингов, после которого брокер запускает ребаланс
	SessionTimeout time.Duration `mapstructure:"session_timeout" json:"session_timeout,omitempty" yaml:"session_timeout,omitempty" toml:"session_timeout,omitempty"`
	// RebalanceTimeout Время, выделяемое воркеру на сдачу партиций при ребалансе группы
	RebalanceTimeout time.Duration `mapstructure:"rebalance_timeout" json:"rebalance_timeout,omitempty" yaml:"rebalance_timeout,omitempty" toml:"rebalance_timeout,omitempty"`
	// ReadTimeout Низкоуровневый таймаут сетевого сокета на чтение
	ReadTimeout time.Duration `mapstructure:"read_timeout" json:"read_timeout,omitempty" yaml:"read_timeout,omitempty" toml:"read_timeout,omitempty"`

	// Тонкие настройки производительности рантайма библиотеки kafka-go

	// MaxAttempts Лимит попыток переподключения до генерации критической ошибки
	MaxAttempts int `mapstructure:"max_attempts" json:"max_attempts,omitempty" yaml:"max_attempts,omitempty" toml:"max_attempts,omitempty"`
	// QueueCapacity Емкость фонового внутреннего канала предвыборки данных из сети
	QueueCapacity int `mapstructure:"queue_capacity" json:"queue_capacity,omitempty" yaml:"queue_capacity,omitempty" toml:"queue_capacity,omitempty"`
	// StartOffset Точка старта при отсутствии сохраненных оффсетов ("first"/"last")
	StartOffset string `mapstructure:"start_offset" json:"start_offset,omitempty" yaml:"start_offset,omitempty" toml:"start_offset,omitempty"`
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

// NewDefaultKafkaReceiverConfig собирает конфигурацию консьюмера по умолчанию с наполнением системными дефолтами.
func NewDefaultKafkaReceiverConfig() *KafkaReceiverConfig {
	return NewKafkaReceiverConfig(
		DefaultKafkaBrokers,
		"",
		"",
		DefaultKafkaReceiverPartition, // Дефолтный маркер отсутствия явной партиции (включает режим GroupID)
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

// Validate выполняет строгую семантическую и математическую валидацию параметров получателя на этапе запуска (Bootstrap Phase).
// Предотвращает запуск консьюмеров с неконсистентными таймаутами сокетов, нарушенными правилами пингов или пустыми топиками.
func (krc *KafkaReceiverConfig) Validate() error {
	if len(krc.Brokers) == 0 {
		return errs.NewConfigValidateError("kafka receiver", "Brokers", "at least one broker address is required", nil)
	}
	if strings.TrimSpace(krc.TargetName) == "" {
		return errs.NewConfigValidateError("kafka receiver", "TargetName", "empty (required parameter for start)", nil)
	}

	hasGroup := strings.TrimSpace(krc.GroupID) != ""
	hasPartition := krc.Partition >= 0 // ИСПРАВЛЕНО: Любое значение >= 0 означает, что партиция явно указана

	// Проверка XOR (Исключающее ИЛИ): должно быть строго заполнено ровно одно из двух
	if !hasGroup && !hasPartition {
		return errs.NewConfigValidateError("kafka receiver", "GroupID/Partition", "either GroupID must be set or Partition must be >= 0", nil)
	}
	if hasGroup && hasPartition {
		return errs.NewConfigValidateError("kafka receiver", "GroupID/Partition", "GroupID and Partition are mutually exclusive options", nil)
	}
	if krc.ConnectTimeout <= 0 {
		return errs.NewConfigValidateError("kafka receiver", "ConnectTimeout", "less than or equal to 0", nil)
	}
	if krc.ShutdownTimeout <= 0 {
		return errs.NewConfigValidateError("kafka receiver", "ShutdownTimeout", "less than or equal to 0", nil)
	}
	if krc.MinBytes <= 0 || krc.MaxBytes <= 0 {
		return errs.NewConfigValidateError("kafka receiver", "MinBytes/MaxBytes", "must be greater than 0", nil)
	}
	// Золотое правило координации Kafka (только для режима группы): сессионный таймаут должен вмещать минимум 3 попытки отправки heartbeat пингов
	if hasGroup && krc.SessionTimeout < krc.HeartbeatInterval*3 {
		return errs.NewConfigValidateError("kafka receiver", "SessionTimeout", "must be at least 3 times greater than heartbeat interval", nil)
	}
	// Защита сокетов от ложных io.EOF ошибок: таймаут чтения обязан перекрывать время холостого ожидания пачки брокером
	if krc.ReadTimeout <= krc.MaxWait {
		return errs.NewConfigValidateError("kafka receiver", "ReadTimeout", "must be strictly greater than MaxWait to prevent socket EOF on low-volume topics", nil)
	}

	// Валидируем точку входа оффсетов, приводя текстовое значение к нижнему регистру
	startOffsetLower := strings.ToLower(strings.TrimSpace(krc.StartOffset))
	if startOffsetLower != "first" && startOffsetLower != "last" {
		return errs.NewConfigValidateError("kafka receiver", "StartOffset", "must be strictly 'first' or 'last'", nil)
	}

	return nil
}

// BuildDefaultKafkaReceiverSettings kafka receiver settings builder
//
// # ВНИМАНИЕ!!!
//
// Сюда не входят:
//
// # TemplateKeyKafkaReceiverTargetName TemplateFlagKafkaReceiverTargetName
//
// # TemplateKeyKafkaReceiverGroupID TemplateFlagKafkaReceiverGroupID
//
// # TemplateKeyKafkaReceiverUsername TemplateFlagKafkaReceiverUsername
//
// TemplateKeyKafkaReceiverPassword TemplateFlagKafkaReceiverPassword
func BuildDefaultKafkaReceiverSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverBrokers, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverBrokers, flagPrefix), DefaultKafkaBrokers, fmt.Sprintf("%s kafka brokers, separated by comma, like: localhost:9092,localhost:9093", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverTargetName, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverTargetName, flagPrefix), "", fmt.Sprintf("%s kafka queue/topic name", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverGroupID, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverGroupID, flagPrefix), "", fmt.Sprintf("%s kafka group id", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverPartition, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverPartition, flagPrefix), DefaultKafkaReceiverPartition, fmt.Sprintf("%s kafka partition", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverConnectTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverConnectTimeout, flagPrefix), DefaultKafkaReceiverConnectTimeout, fmt.Sprintf("%s kafka connect timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverShutdownTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverShutdownTimeout, flagPrefix), DefaultKafkaReceiverShutdownTimeout, fmt.Sprintf("%s kafka shutdown timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverMinBytes, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverMinBytes, flagPrefix), DefaultKafkaReceiverMinBytes, fmt.Sprintf("%s kafka min bytes", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverMaxBytes, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverMaxBytes, flagPrefix), DefaultKafkaReceiverMaxBytes, fmt.Sprintf("%s kafka max bytes", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverMaxWait, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverMaxWait, flagPrefix), DefaultKafkaReceiverMaxWait, fmt.Sprintf("%s kafka max wait", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverUsername, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverUsername, flagPrefix), "", fmt.Sprintf("%s kafka username", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverPassword, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverPassword, keyPrefix), "", fmt.Sprintf("%s kafka password", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverHeartbeatInterval, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverHeartbeatInterval, flagPrefix), DefaultKafkaReceiverHeartbeatInterval, fmt.Sprintf("%s kafka heartbeat interval", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverSessionTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverSessionTimeout, flagPrefix), DefaultKafkaReceiverSessionTimeout, fmt.Sprintf("%s kafka session timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverRebalanceTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverRebalanceTimeout, flagPrefix), DefaultKafkaReceiverRebalanceTimeout, fmt.Sprintf("%s kafka rebalance timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverReadTimeout, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverReadTimeout, flagPrefix), DefaultKafkaReceiverReadTimeout, fmt.Sprintf("%s kafka read timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverMaxAttempts, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverMaxAttempts, flagPrefix), DefaultKafkaReceiverMaxAttempts, fmt.Sprintf("%s kafka max attempts", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverQueueCapacity, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverQueueCapacity, flagPrefix), DefaultKafkaReceiverQueueCapacity, fmt.Sprintf("%s kafka queue capacity", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyKafkaReceiverStartOffset, keyPrefix), fmt.Sprintf(TemplateFlagKafkaReceiverStartOffset, flagPrefix), DefaultKafkaReceiverStartOffset, fmt.Sprintf("%s kafka start offset", flagPrefix)),
	}
}
