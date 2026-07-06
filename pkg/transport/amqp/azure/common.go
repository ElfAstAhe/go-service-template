package azure

import (
	"context"

	"github.com/Azure/go-amqp"
)

// amqpSenderLink описывает методы встроенного отправителя библиотеки Azure AMQP,
// которые нам нужны для управления его жизненным циклом.
type amqpSenderLink interface {
	Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error
	Close(ctx context.Context) error
}
