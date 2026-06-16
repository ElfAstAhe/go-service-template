// Package pubsub specification of Publisher/Subscriber observer template
package pubsub

import (
	"context"
)

type Publisher[T any] interface {
	Register(Observer[T])
	Unregister(Observer[T])
	Notify(context.Context, T)
}

type Observer[T any] interface {
	GetName() string
	OnNotify(context.Context, T) error
}
