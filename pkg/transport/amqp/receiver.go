package amqp

import (
	"context"
)

// Receiver описывает чистый контракт для получения сообщений из конкретной очереди/топика AMQP 1.0.
//
//	ReceiveOpts — опции получателя конкретного кадра сообщения
//	MsgHeader   — заголовок сообщения (дженерик)
type Receiver[ReceiveOpts any, MsgHeader any] interface {
	// Receive блокирует поток до тех пор, пока из настроенной очереди/топика
	// не прилетит новое сообщение, либо пока не отменится контекст.
	Receive(ctx context.Context, receiveOpts ReceiveOpts) (*Message[MsgHeader], error)

	// Accept подтверждает брокеру успешную обработку сообщения. Message удаляется из очереди.
	Accept(ctx context.Context, msg *Message[MsgHeader]) error

	// Reject сообщает о критической ошибке обработки (например, битый JSON).
	// Брокер уводит сообщение в DLA (Dead Letter Address), защищая воркеров от бесконечного цикла падений.
	Reject(ctx context.Context, msg *Message[MsgHeader], err error) error

	// Release сообщает о временной ошибке (например, упала БД).
	// Брокер возвращает сообщение обратно в очередь для повторной обработки.
	Release(ctx context.Context, msg *Message[MsgHeader]) error

	// Close мягко закрывает слушающий линк, не прерывая общую сессию коннектора.
	Close(ctx context.Context) error

	GetTargetName() string
}
