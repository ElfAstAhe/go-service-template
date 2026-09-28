package metrics

const (
	// StatusSuccess определяет строковый маркер успешного завершения операции (HTTP 2xx, Kafka Ack) для лейблов метрик.
	StatusSuccess = "success"
	// StatusFail определяет строковый маркер сбоя или ошибки выполнения операции для лейблов метрик.
	StatusFail = "fail"
)
