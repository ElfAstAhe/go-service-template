package broker

import (
	"context"
)

// Receiver описывает чистый контракт для получения сообщений из конкретной очереди/топика AMQP 1.0.
type Receiver interface {
	// Receive блокирует поток до тех пор, пока из настроенной очереди/топика
	// не прилетит новое сообщение, либо пока не отменится контекст.
	Receive(ctx context.Context) (Message, error)

	// Accept подтверждает брокеру успешную обработку сообщения. Message удаляется из очереди.
	Accept(ctx context.Context, msg Message) error

	// Reject сообщает о критической ошибке обработки (например, битый JSON).
	// Брокер уводит сообщение в DLA (Dead Letter Address), защищая воркеров от бесконечного цикла падений.
	Reject(ctx context.Context, msg Message, err error) error

	// Release сообщает о временной ошибке (например, упала БД).
	// Брокер возвращает сообщение обратно в очередь для повторной обработки.
	Release(ctx context.Context, msg Message) error

	// Close мягко закрывает слушающий линк, не прерывая общую сессию коннектора.
	Close(ctx context.Context) error

	// GetTargetName информация о топике/очереди
	GetTargetName() string

	// Stats возвращает строго типизированную структуру метрик. No more map[string]any!
	Stats() ReceiverStats
}
