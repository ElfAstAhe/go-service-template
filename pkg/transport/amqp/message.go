package amqp

// Message — наш собственный независимый контейнер для AMQP 1.0 сообщения.
//
//	TargetName - topic/queue при получении
//	Header - заголовок при отправке
//	Payload - содержимое
//	Properties - свойства
type Message[Header any] struct {
	TargetName string
	Header     Header
	Payload    []byte         // Сырые байты (например, JSON)
	Properties map[string]any // Заголовки / Метаданные (для TraceID, etc.)
}
