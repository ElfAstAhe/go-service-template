package amqp

import (
	"context"
)

// Connector interface for connector standardization
type Connector[Connection any] interface {
	Open(ctx context.Context) error
	Close(ctx context.Context) error
	GetConnection(ctx context.Context) (Connection, error)
	IsConnected() bool
}
