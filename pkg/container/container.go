package container

import (
	"context"
)

// Container интерфейс контейнера
type Container interface {
	GetName() string

	Init(ctx context.Context) error
	Close(ctx context.Context) error

	RegisterInstance(name string, instance any) error
	UnregisterInstance(name string) error

	GetInstance(name string) (any, error)
	IsRegistered(name string) bool

	AllInstances() map[string]any
}
