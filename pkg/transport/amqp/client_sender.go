package amqp

import (
	"context"
)

type ClientSender interface {
	Publish(ctx context.Context, address string, msg *Message) error
}
