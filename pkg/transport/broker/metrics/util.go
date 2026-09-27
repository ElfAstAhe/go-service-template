package metrics

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/metrics"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
)

// UpdateReceiverMetrics — атомарный мост, транслирующий плоский Snapshot из Stats() в Prometheus.
// Метод полностью безопасен: защищен от пустых полей и невалидных кастомных значений.
func UpdateReceiverMetrics(stats pkgamqp.ReceiverStats) {
	// 1. Обновляем COMMON блок (заполняется всегда)
	metrics.ReceiverMessagesTotal.WithLabelValues(stats.BrokerType, stats.TargetName, stats.Status).Set(float64(stats.TotalMessages))
	metrics.ReceiverErrorsTotal.WithLabelValues(stats.BrokerType, stats.TargetName).Set(float64(stats.TotalErrors))

	// 2. Обновляем KAFKA блок (точечный предохранитель)
	if stats.BrokerType == "kafka" {
		if stats.Lag >= 0 {
			metrics.ReceiverKafkaLag.WithLabelValues(stats.BrokerType, stats.TargetName).Set(float64(stats.Lag))
		}
		if stats.QueueLength >= 0 {
			metrics.ReceiverKafkaQueueLength.WithLabelValues(stats.BrokerType, stats.TargetName).Set(float64(stats.QueueLength))
		}
	}

	// 3. Обновляем AMQP блок (точечный предохранитель)
	if stats.BrokerType == "azure-amqp" || stats.BrokerType == "artemis" || stats.BrokerType == "rabbitmq" {
		if stats.PrefetchCount >= 0 {
			metrics.ReceiverAMQPPrefetchCount.WithLabelValues(stats.BrokerType, stats.TargetName).Set(float64(stats.PrefetchCount))
		}
		if stats.ConsumerCount >= 0 {
			metrics.ReceiverAMQPConsumerCount.WithLabelValues(stats.BrokerType, stats.TargetName).Set(float64(stats.ConsumerCount))
		}
	}
}

// ====================================================================
// Вспомогательные хелперы для инкремента метрик из рантайма
// ====================================================================

// TracePublishDuration фиксирует время сетевого пуша (вызывать через defer в начале Publish)
func TracePublishDuration(brokerType, targetName string, startTime time.Time) {
	duration := time.Since(startTime).Seconds()
	metrics.SenderPublishDurationHistogram.WithLabelValues(brokerType, targetName).Observe(duration)
}

// IncPublishSuccess увеличивает счетчик успешных доставок
func IncPublishSuccess(brokerType, targetName string) {
	metrics.SenderPushedMessagesTotal.WithLabelValues(brokerType, targetName).Inc()
}

// IncPublishRetry увеличивает счетчик ретраев
func IncPublishRetry(brokerType, targetName string) {
	metrics.SenderPublishRetriesTotal.WithLabelValues(brokerType, targetName).Inc()
}

// IncPublishError фиксирует фатальную ошибку
func IncPublishError(brokerType, targetName, errorType string) {
	metrics.SenderPublishErrorsTotal.WithLabelValues(brokerType, targetName, errorType).Inc()
}
