package kafka

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/segmentio/kafka-go"
)

// Константы дефолтов для внутренней защиты рантайм-компонента (независимые от пакета config)
const (
	defaultReceiverConnectTimeout   = 10 * time.Second
	defaultReceiverShutdownTimeout  = 15 * time.Second
	defaultReceiverMinBytes         = 1024
	defaultReceiverMaxBytes         = 10e6 // 10 MB
	defaultReceiverMaxWait          = 500 * time.Millisecond
	defaultReceiverHeartbeat        = 3 * time.Second
	defaultReceiverSessionTimeout   = 30 * time.Second
	defaultReceiverRebalance        = 60 * time.Second
	defaultReceiverReadBatchTimeout = 10 * time.Second
	defaultReceiverMaxAttempts      = 3
	defaultReceiverQueueCapacity    = 100
	defaultReceiverStartOffset      = "first"
)

// defaultReceiverBrokers хранит локальный адрес по умолчанию
var defaultReceiverBrokers = []string{"localhost:9092"}

// ReceiverOption определяет функциональный тип для конфигурации опций получателя Kafka (Fluent API).
type ReceiverOption func(*ReceiverOptions)

// ReceiverOptions содержит параметры рантайма, необходимые для безопасной сборки и работы консьюмера Kafka.
type ReceiverOptions struct {
	ClientID          string
	Brokers           []string            // Прямой список хостов брокеров (вместо старого Connector)
	TargetName        string              // Имя топика (Topic)
	GroupID           string              // Идентификатор Consumer Group
	Partition         int                 // идентификатор Direct consumer
	ReaderConf        *kafka.ReaderConfig // Дополнительные низкоуровневые кастомные опции библиотеки kafka-go
	TLS               *tls.Config
	ConnectTimeout    time.Duration
	ShutdownTimeout   time.Duration
	HeartbeatInterval time.Duration
	SessionTimeout    time.Duration
	RebalanceTimeout  time.Duration // Таймаут на сдачу оффсетов при ребалансе
	ReadBatchTimeout  time.Duration
	Logger            logger.Logger
	MinBytes          int
	MaxBytes          int
	MaxWait           time.Duration
	Username          string
	Password          string
	MaxAttempts       int    // Лимит попыток подключения библиотеки
	QueueCapacity     int    // Размер фонового буфера сообщений
	StartOffset       string // Точка старта ("first" или "last")
}

// NewReceiverOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewReceiverOptions() *ReceiverOptions {
	return &ReceiverOptions{
		Brokers:           defaultReceiverBrokers,
		GroupID:           "",
		Partition:         -1,
		ConnectTimeout:    defaultReceiverConnectTimeout,
		ShutdownTimeout:   defaultReceiverShutdownTimeout,
		MinBytes:          defaultReceiverMinBytes,
		MaxBytes:          defaultReceiverMaxBytes,
		MaxWait:           defaultReceiverMaxWait,
		HeartbeatInterval: defaultReceiverHeartbeat,
		SessionTimeout:    defaultReceiverSessionTimeout,
		RebalanceTimeout:  defaultReceiverRebalance,
		ReadBatchTimeout:  defaultReceiverReadBatchTimeout,
		MaxAttempts:       defaultReceiverMaxAttempts,
		QueueCapacity:     defaultReceiverQueueCapacity,
		StartOffset:       defaultReceiverStartOffset,
	}
}

// Validate проверяет корректность абсолютно всех опций рантайма перед сборкой Receiver.
// Защищает приложение от паник библиотеки kafka-go и некорректного поведения консьюмера.
func (ro *ReceiverOptions) Validate() error {
	// 1. Проверка базовых инфраструктурных полей
	if len(ro.Brokers) == 0 {
		return errs.NewTlCommonError("Validate", "at least one broker address is required", nil)
	}
	if strings.TrimSpace(ro.TargetName) == "" {
		return errs.NewTlCommonError("Validate", "target name (topic) cannot be empty", nil)
	}
	if ro.Logger == nil {
		return errs.NewTlCommonError("Validate", "logger is required and cannot be nil", nil)
	}

	// ИСПРАВЛЕНО: Кросс-валидация режимов (Consumer Group vs Direct Partition Assignment)
	hasGroup := strings.TrimSpace(ro.GroupID) != ""
	hasPartition := ro.Partition >= 0

	if !hasGroup && !hasPartition {
		return errs.NewTlCommonError("Validate", "either GroupID must be set or Partition must be >= 0 for reader initialization", nil)
	}
	if hasGroup && hasPartition {
		return errs.NewTlCommonError("Validate", "GroupID and Partition are mutually exclusive options and cannot be used together", nil)
	}

	// 2. Проверка базовых сетевых таймаутов
	if ro.ConnectTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "ConnectTimeout must be greater than 0", nil)
	}
	if ro.ShutdownTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "ShutdownTimeout must be greater than 0", nil)
	}

	// 3. Проверка параметров вычитки (Fetch bounds)
	if ro.MinBytes <= 0 {
		return errs.NewTlCommonError("Validate", "MinBytes must be greater than 0", nil)
	}
	if ro.MaxBytes <= 0 {
		return errs.NewTlCommonError("Validate", "MaxBytes must be greater than 0", nil)
	}
	if ro.MaxBytes < ro.MinBytes {
		return errs.NewTlCommonError("Validate", "MaxBytes cannot be less than MinBytes", nil)
	}
	if ro.MaxWait <= 0 {
		return errs.NewTlCommonError("Validate", "MaxWait (broker poll wait) must be greater than 0", nil)
	}

	// 4. Проверка таймаутов координации группы (ИСПРАВЛЕНО: проверяем только для regime Consumer Group)
	if hasGroup {
		if ro.HeartbeatInterval <= 0 {
			return errs.NewTlCommonError("Validate", "HeartbeatInterval must be greater than 0", nil)
		}
		if ro.SessionTimeout <= 0 {
			return errs.NewTlCommonError("Validate", "SessionTimeout must be greater than 0", nil)
		}
		if ro.RebalanceTimeout <= 0 {
			return errs.NewTlCommonError("Validate", "RebalanceTimeout must be greater than 0", nil)
		}
		// Золотое правило Kafka: сессия должна пережить как минимум 3 пропущенных пинга
		if ro.SessionTimeout < ro.HeartbeatInterval*3 {
			errMsg := fmt.Sprintf("session timeout (%v) must be at least 3 times greater than heartbeat interval (%v)", ro.SessionTimeout, ro.HeartbeatInterval)
			return errs.NewTlCommonError("Validate", errMsg, nil)
		}
	}

	// 5. Проверка сетевого таймаута сокета на чтение батча
	if ro.ReadBatchTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "ReadBatchTimeout must be greater than 0", nil)
	}
	if ro.ReadBatchTimeout <= ro.MaxWait {
		errMsg := fmt.Sprintf("ReadBatchTimeout (%v) must be strictly greater than MaxWait (%v) to prevent socket EOF on low-volume topics", ro.ReadBatchTimeout, ro.MaxWait)
		return errs.NewTlCommonError("Validate", errMsg, nil)
	}

	// 6. Проверка параметров производительности
	if ro.MaxAttempts <= 0 {
		return errs.NewTlCommonError("Validate", "MaxAttempts (reconnect attempts) must be at least 1", nil)
	}
	if ro.QueueCapacity <= 0 {
		return errs.NewTlCommonError("Validate", "QueueCapacity (internal pre-fetch buffer) must be greater than 0", nil)
	}

	// 7. Валидация политики точки старта
	startOffsetLower := strings.ToLower(strings.TrimSpace(ro.StartOffset))
	if startOffsetLower != "first" && startOffsetLower != "last" {
		errMsg := fmt.Sprintf("invalid StartOffset '%s', allowed values are strictly 'first' or 'last'", ro.StartOffset)
		return errs.NewTlCommonError("Validate", errMsg, nil)
	}

	return nil
}

// WithReceiverClientID задает уникальный строковый ClientID идентификатор текущего консьюмера для Kafka-кластера.
func WithReceiverClientID(clientID string) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.ClientID = clientID
	}
}

// WithReceiverBrokers инжектирует список целевых сетевых адресов брокеров Kafka.
func WithReceiverBrokers(brokers []string) ReceiverOption {
	return func(ro *ReceiverOptions) { ro.Brokers = brokers }
}

// WithReceiverTargetName задает строковое имя целевого топика Kafka, из которого будет производиться чтение.
func WithReceiverTargetName(targetName string) ReceiverOption {
	return func(ro *ReceiverOptions) { ro.TargetName = targetName }
}

// WithReceiverGroupID инжектирует идентификатор Consumer Group для распределенной вычитки топика.
func WithReceiverGroupID(groupID string) ReceiverOption {
	return func(ro *ReceiverOptions) { ro.GroupID = groupID }
}

// WithReceiverPartition переключает консьюмер на чтение конкретной физической партиции топика напрямую.
func WithReceiverPartition(partition int) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.Partition = partition
	}
}

// WithKafkaReaderConfig позволяет передать низкоуровневую структуру параметров kafka.ReaderConfig напрямую.
func WithKafkaReaderConfig(cfg *kafka.ReaderConfig) ReceiverOption {
	return func(ro *ReceiverOptions) { ro.ReaderConf = cfg }
}

// WithKafkaReceiverTLS инжектирует криптографические TLS-параметры шифрования сетевого трафика.
func WithKafkaReceiverTLS(tls *tls.Config) ReceiverOption {
	return func(ro *ReceiverOptions) { ro.TLS = tls }
}

// WithReceiverConnectTimeout задает лимит времени на установление первичного сетевого соединения с кластером.
func WithReceiverConnectTimeout(timeout time.Duration) ReceiverOption {
	return func(ro *ReceiverOptions) { ro.ConnectTimeout = timeout }
}

// WithReceiverShutdownTimeout задает таймаут на плавное освобождение ресурсов и сдачу оффсетов при выходе.
func WithReceiverShutdownTimeout(timeout time.Duration) ReceiverOption {
	return func(ro *ReceiverOptions) { ro.ShutdownTimeout = timeout }
}

// WithReceiverFetchBounds настраивает границы размеров пакетов и время ожидания накопления данных брокером.
func WithReceiverFetchBounds(minBytes, maxBytes int, maxWait time.Duration) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.MinBytes = minBytes
		ro.MaxBytes = maxBytes
		ro.MaxWait = maxWait
	}
}

// WithReceiverLogger инжектирует структурированный логгер фреймворка для аудита рантайма получателя.
func WithReceiverLogger(log logger.Logger) ReceiverOption {
	return func(ro *ReceiverOptions) { ro.Logger = log }
}

// WithReceiverSecurity задает строковые параметры авторизации SASL Plain для доступа к защищенным топикам.
func WithReceiverSecurity(username, password string) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.Username = username
		ro.Password = password
	}
}

// WithReceiverGroupTimeouts гранулярно настраивает параметры сессий, фоновых пингов и ребалансировок Consumer Group.
func WithReceiverGroupTimeouts(heartbeat, session, rebalance, readBatch time.Duration) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.HeartbeatInterval = heartbeat
		ro.SessionTimeout = session
		ro.RebalanceTimeout = rebalance
		ro.ReadBatchTimeout = readBatch
	}
}

// WithReceiverRuntimePerformance конфигурирует параметры внутренней производительности, лимиты ретраев и размер фонового буфера.
func WithReceiverRuntimePerformance(maxAttempts, queueCapacity int, startOffset string) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.MaxAttempts = maxAttempts
		ro.QueueCapacity = queueCapacity
		ro.StartOffset = startOffset
	}
}
