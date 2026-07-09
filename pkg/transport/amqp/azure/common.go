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
