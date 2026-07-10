package azure

import (
	"context"

	"github.com/Azure/go-amqp"
)

// AmqpSenderLink описывает методы встроенного отправителя библиотеки Azure AMQP,
// которые нам нужны для управления его жизненным циклом.
type AmqpSenderLink interface {
	Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error
	Close(ctx context.Context) error
}

// AmqpReceiverLink описывает методы встроенного получателя Azure AMQP,
// необходимые для чтения, подтверждения и закрытия линка.
type AmqpReceiverLink interface {
	Receive(ctx context.Context, opts *amqp.ReceiveOptions) (*amqp.Message, error)
	AcceptMessage(ctx context.Context, msg *amqp.Message) error
	RejectMessage(ctx context.Context, msg *amqp.Message, err *amqp.Error) error
	ReleaseMessage(ctx context.Context, msg *amqp.Message) error
	Close(ctx context.Context) error
}
