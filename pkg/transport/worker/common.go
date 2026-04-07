package worker

import (
	"context"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type JobHandler[D any] func(ctx context.Context, workerIndex int, data D) error

type DispatcherDataProvider[D comparable] func(ctx context.Context, eventTime time.Time) ([]D, error)

type CommonWorker interface {
	Start(ctx context.Context) error
	Stop(stopTimeOut time.Duration) error

	GetName() string
	GetContext() context.Context
	GetContextCancel() context.CancelFunc
	GetLogger() logger.Logger
	GetWaitGroup() *sync.WaitGroup
	IsRunning() bool
}
