package container

import (
	"context"
)

type SimpleCloser interface {
	Close() error
}

type ContextCloser interface {
	Close(ctx context.Context) error
}

type closeInstance struct {
	Name     string
	Instance any
}

func newCloseInstance(name string, instance any) *closeInstance {
	return &closeInstance{
		Name:     name,
		Instance: instance,
	}
}
