package amqp

import (
	"context"
)

// Connector interface for connector standardization
type Connector[Connection any] interface {
	Close(ctx context.Context) error
	GetConnection(ctx context.Context) (Connection, error)
	Invalidate(err error)
}
