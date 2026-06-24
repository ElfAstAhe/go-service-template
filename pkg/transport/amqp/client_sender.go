package amqp

import (
	"context"
)

type ClientSender[O any] interface {
	Publish(ctx context.Context, targetName string, msg *Message, sendOpts O) error
	Close(ctx context.Context) error
}
