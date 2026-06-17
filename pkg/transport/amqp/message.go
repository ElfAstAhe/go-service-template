package amqp

// Message — наш собственный независимый контейнер для AMQP 1.0 сообщения.
type Message struct {
	Payload    []byte         // Сырые байты (например, JSON)
	Properties map[string]any // Заголовки / Метаданные (для TraceID, etc.)
}
