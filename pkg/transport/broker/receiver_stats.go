package broker

import (
	"time"
)

// ReceiverStats — унифицированный enterprise-контракт метрик для любого типа получателя во фреймворке.
// Использует JSON-теги с omitempty для сокрытия специфичных полей в зависимости от активного брокера.
type ReceiverStats struct {
	// ====================================================================
	// COMMON Specific (Общие метрики для всех типов брокеров)
	// ====================================================================
	BrokerType    string    `json:"broker_type"`    // Тип брокера: "kafka", "artemis", "rabbitmq"
	TargetName    string    `json:"target_name"`    // Имя топика брокера или очереди
	Status        string    `json:"status"`         // Текущий статус: "connected", "disconnected", "rebalancing"
	TotalMessages uint64    `json:"total_messages"` // Общее количество успешно вычитанных сообщений (Counter)
	TotalErrors   uint64    `json:"total_errors"`   // Общее количество ошибок ввода-вывода (Counter)
	Lag           int64     `json:"lag"`            // Отставание (Gauge). Количество оставшихся сообщений в брокере.
	ConnectedAt   time.Time `json:"connected_at"`   // Время последней успешной установки соединения

	// ====================================================================
	// KAFKA Specific (Метрики, применимые только к архитектуре Kafka)
	// ====================================================================
	Partition     string `json:"partition,omitempty"`      // Активная партиция кластера (или список партиций)
	Offset        int64  `json:"offset,omitempty"`         // Текущее зафиксированное смещение (Offset) ридера
	QueueLength   int64  `json:"queue_length,omitempty"`   // Текущая заполненность внутреннего буфера предвыборки в памяти
	QueueCapacity int64  `json:"queue_capacity,omitempty"` // Максимальная емкость внутреннего буфера (QueueCapacity)

	// ====================================================================
	// AMQP Specific (Метрики, применимые только к классическим MQ: Artemis, RabbitMQ)
	// ====================================================================
	ConsumerCount int64 `json:"consumer_count,omitempty"` // Количество активных консьюмеров на данной очереди
	PrefetchCount int64 `json:"prefetch_count,omitempty"` // Настройка лимита предвыборки (QoS Prefetch) на канале
}
