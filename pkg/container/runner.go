package container

import (
	"context"
	"time"
)

type Runner interface {
	GetName() string
	Start(ctx context.Context) error
	Stop(timeout time.Duration) error
	IsRunning() bool
}
