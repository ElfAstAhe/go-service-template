package amqp

import (
	"context"
)

// ClientReceiver описывает чистый контракт для получения сообщений из брокера AMQP 1.0.
// Он намеренно изолирован от сендера, чтобы сервисы импортировали только то, что им нужно.
//
//	ReceiveOpts - опции получателя
//	MsgHeader - заголовок сообщения
type ClientReceiver[ReceiveOpts any, MsgHeader any] interface {
	// Receive блокирует поток до тех пор, пока из указанной очереди/топика (targetName)
	// не прилетит новое сообщение, либо пока не отменится контекст.
	Receive(ctx context.Context, targetName string, receiveOpts ReceiveOpts) (*Message[MsgHeader], error)

	// Accept подтверждает брокеру успешную обработку сообщения. Message удаляется из очереди.
	Accept(ctx context.Context, msg *Message[MsgHeader]) error

	// Reject сообщает о критической ошибке обработки (например, битый JSON).
	// Брокер уводит сообщение в DLA (Dead Letter Address), защищая воркеров от бесконечного цикла падений.
	Reject(ctx context.Context, msg *Message[MsgHeader], err error) error

	// Release сообщает о временной ошибке (например, упала БД).
	// Брокер возвращает сообщение обратно в очередь для повторной обработки.
	Release(ctx context.Context, msg *Message[MsgHeader]) error

	// Close мягко закрывает слушающие линки, сессию и соединение с брокером.
	Close(ctx context.Context) error

	GetTargetNames() []string
}
