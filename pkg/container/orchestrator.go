package container

import (
	"context"
)

type Orchestrator interface {
	Init(ctx context.Context) error
	Close(ctx context.Context) error

	Register(container Container) error
	Unregister(name string) error

	GetContainer(name string) (Container, error)
	HasContainer(name string) bool

	AllContainers() []Container

	GetRunners() ([]Runner, error)
}
