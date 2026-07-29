package amqp

// Message — наш собственный независимый контейнер для AMQP 1.0 сообщения.
type Message interface {
	GetTargetName() string
	GetPayload() []byte
	GetProperties() map[string]any
	ExtractOriginalMessage() (any, error)
}
