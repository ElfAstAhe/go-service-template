package broker

// Message определяет унифицированный, абстрактный контракт транспортного сообщения фреймворка (Unified Message Wrapper).
//
// Выступает независимым полиморфным контейнером-оберткой над специфичными структурами брокеров очередей (Kafka, AMQP, RabbitMQ).
// Полностью изолирует UseCase-слой и фоновых воркеров от низкоуровневых типов данных конкретных шин трафика.
type Message interface {
	// GetTargetName возвращает имя целевого назначения сообщения (например, название топика Kafka или очереди RabbitMQ).
	GetTargetName() string

	// GetPayload возвращает сырой массив байт (бинарное тело payload) сообщения для последующей десериализации.
	GetPayload() []byte

	// GetProperties возвращает карту метаданных, заголовков и служебных атрибутов (Headers/Properties) сообщения.
	GetProperties() map[string]any

	// ExtractOriginalMessage извлекает исходный, вендоро-специфичный объект сообщения (например, *kafka.Message или amqp.Delivery).
	// Используется низкоуровневыми инфраструктурными хелперами для ручной фиксации оффсетов (Manual Ack/Commit).
	ExtractOriginalMessage() (any, error)
}
