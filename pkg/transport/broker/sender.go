package broker

import (
	"context"
)

// Sender предназначен для работы по схеме 1 к 1.
// Он жестко завязан на одну конкретную очередь/топик, переданную в конструктор.
type Sender interface {
	Publish(ctx context.Context, msg Message) error
	Close(ctx context.Context) error

	GetTargetName() string
}
