package amqp

import (
	"context"
)

type ClientSender interface {
	Publish(ctx context.Context, targetName string, msg *Message) error
	Close(ctx context.Context) error
}
