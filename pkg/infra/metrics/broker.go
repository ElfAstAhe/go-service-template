package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Названия метрик и их описание приведены к индустриальному стандарту OpenTelemetry/Prometheus.
// Используем promauto для автоматической регистрации метрик в DefaultRegisterer.
var (
	// ====================================================================
	// COMMON Metrics (Общие метрики для всех типов брокеров)
	// ====================================================================

	// ReceiverMessagesTotal фиксирует общее количество успешно прочитанных и обработанных сообщений.
	ReceiverMessagesTotal *prometheus.GaugeVec

	// ReceiverErrorsTotal считает сетевые сбои, таймауты сокетов или ошибки десериализации payload получателя.
	ReceiverErrorsTotal *prometheus.GaugeVec

	// ====================================================================
	// KAFKA Specific Metrics (Только для архитектуры Kafka)
	// ====================================================================

	// ReceiverKafkaLag показывает текущее отставание консьюмера (количество невычитанных сообщений в партициях Kafka).
	ReceiverKafkaLag *prometheus.GaugeVec

	// ReceiverKafkaQueueLength трекает заполненность фонового буфера предвыборки в памяти консьюмера kafka-go.
	ReceiverKafkaQueueLength *prometheus.GaugeVec

	// ====================================================================
	// AMQP Specific Metrics (Для классических MQ: Azure Service Bus, Artemis)
	// ====================================================================

	// ReceiverAMQPPrefetchCount показывает текущий лимит предвыборки (Flow Control Credits) AMQP линка.
	ReceiverAMQPPrefetchCount *prometheus.GaugeVec

	// ReceiverAMQPConsumerCount считает количество активных конкурирующих консьюмеров на очереди брокера.
	ReceiverAMQPConsumerCount *prometheus.GaugeVec
)

// Названия метрик и их описание приведены к индустриальному стандарту OpenTelemetry/Prometheus.
// Используем promauto для автоматической регистрации метрик в DefaultRegisterer.
var (
	// ====================================================================
	// SENDER COMMON Metrics (Общие метрики для всех продюсеров)
	// ====================================================================

	// SenderPushedMessagesTotal фиксирует общее количество вызовов метода Publish.
	SenderPushedMessagesTotal *prometheus.CounterVec

	// SenderPublishErrorsTotal считает критические сбои отправки (когда ретраи исчерпаны).
	SenderPublishErrorsTotal *prometheus.CounterVec

	// SenderPublishRetriesTotal трекает стабильность сети (сколько раз сработал наш бэкофф с джиттером).
	SenderPublishRetriesTotal *prometheus.CounterVec

	// SenderPublishDurationHistogram замеряет скорость ответа брокера (Latency) при публикации в секундах.
	SenderPublishDurationHistogram *prometheus.HistogramVec

	// ====================================================================
	// SENDER KAFKA Specific Metrics (Только для батчинга Kafka)
	// ====================================================================

	// SenderKafkaBatchSize показывает, насколько плотно забиваются пачки (батчи) перед пушем в сеть.
	SenderKafkaBatchSize *prometheus.GaugeVec
)

// InitBrokerReceiverMetrics инициализирует общие векторы метрик для подсистем вычитки сообщений (Consumers).
func InitBrokerReceiverMetrics() {
	ReceiverMessagesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "receiver_messages_total",
		Help: "Общее количество успешно вычитанных и обработанных сообщений фреймворком.",
	}, []string{"broker_type", "target_name", "status"})
	ReceiverErrorsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "receiver_errors_total",
		Help: "Общее количество сетевых сбоев, таймаутов сокетов или ошибок десериализации payload.",
	}, []string{"broker_type", "target_name"})
}

// InitBrokerReceiverAMQPMetrics инициализирует специфичные для протокола AMQP 0-9-1 / 1.0 метрики получателя.
func InitBrokerReceiverAMQPMetrics() {
	ReceiverAMQPPrefetchCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "receiver_amqp_prefetch_count",
		Help: "Текущий лимит предвыборки (Flow Control Credits), выданный активному линку AMQP.",
	}, []string{"broker_type", "target_name"})
	ReceiverAMQPConsumerCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "receiver_amqp_consumer_count",
		Help: "Количество активных конкурирующих консьюмеров на данной очереди брокера.",
	}, []string{"broker_type", "target_name"})
}

// InitBrokerReceiverKafkaMetrics инициализирует специфичные для Apache Kafka метрики отставания и емкости получателя.
func InitBrokerReceiverKafkaMetrics() {
	ReceiverKafkaLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "receiver_kafka_lag",
		Help: "Текущее отставание консьюмера (количество невычитанных сообщений в партициях Kafka).",
	}, []string{"broker_type", "target_name"})
	ReceiverKafkaQueueLength = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "receiver_kafka_queue_length",
		Help: "Текущая заполненность фонового буфера предвыборки в памяти консьюмера kafka-go.",
	}, []string{"broker_type", "target_name"})
}

// InitBrokerSenderMetrics инициализирует сквозные векторы метрик для подсистем публикации сообщений (Producers).
func InitBrokerSenderMetrics() {
	SenderPushedMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sender_messages_total",
		Help: "Общее количество успешно отправленных сообщений в брокер.",
	}, []string{"broker_type", "target_name"})
	SenderPublishErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sender_errors_total",
		Help: "Общее количество фатальных ошибок отправки сообщений.",
	}, []string{"broker_type", "target_name", "error_type"})
	SenderPublishRetriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sender_retries_total",
		Help: "Количество повторных попыток (retries) при временных сетевых сбоях брокера.",
	}, []string{"broker_type", "target_name"})
	SenderPublishDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sender_publish_duration_seconds",
		Help:    "Гистограмма времени задержки (Latency) при публикации сообщения.",
		Buckets: []float64{0.001, 0.005, 0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0, 2.5, 5.0}, // Базовые бакеты от 1мс до 5с
	}, []string{"broker_type", "target_name"})
}

// InitBrokerSenderKafkaMetrics инициализирует специфичные метрики агрегации пачек для продюсера Apache Kafka.
func InitBrokerSenderKafkaMetrics() {
	SenderKafkaBatchSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sender_kafka_batch_messages_count",
		Help: "Текущий размер сформированной пачки (батча) сообщений перед отправкой в сеть Kafka.",
	}, []string{"broker_type", "target_name"})
}
