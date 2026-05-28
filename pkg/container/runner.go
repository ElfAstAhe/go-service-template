package container

import (
	"context"
)

type Runner interface {
	GetName() string
	Start(ctx context.Context) error
	Stop(stopCtx context.Context) error
	IsRunning() bool
}
