package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type BasePoolConfig struct {
	WorkerCount     int
	DataCapacity    int
	CompleteProcess bool
}

func NewBasePoolConfig(workerCount, dataCapacity int, completeProcess bool) *BasePoolConfig {
	return &BasePoolConfig{
		WorkerCount:     workerCount,
		DataCapacity:    dataCapacity,
		CompleteProcess: completeProcess,
	}
}

type BasePool[D any] struct {
	name       string
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	dataChan   chan D
	jobHandler JobHandler[D]
	config     *BasePoolConfig
	log        logger.Logger
	running    *atomic.Bool
}

var _ Pool[string] = (*BasePool[string])(nil)
var _ CommonWorker = (*BasePool[string])(nil)

func NewBasePool[D any](
	name string,
	config *BasePoolConfig,
	jobHandler JobHandler[D],
	logger logger.Logger,
) *BasePool[D] {
	res := &BasePool[D]{
		name:       name,
		config:     config,
		jobHandler: jobHandler,
		log:        logger.GetLogger(name),
		running:    new(atomic.Bool),
	}
	res.running.Store(false)

	return res
}

func (bp *BasePool[D]) Start(ctx context.Context) error {
	if !bp.running.CompareAndSwap(false, true) {
		return errs.NewCommonError(fmt.Sprintf("worker pool %s already started", bp.GetName()), nil)
	}

	bp.GetLogger().Debugf("worker pool %s starting", bp.GetName())
	defer bp.GetLogger().Debugf("worker pool %s started", bp.GetName())

	// context
	bp.ctx, bp.cancel = context.WithCancel(ctx)
	// data
	bp.dataChan = make(chan D, bp.GetConfig().DataCapacity)
	// workers
	for i := 0; i < bp.GetConfig().WorkerCount; i++ {
		bp.GetWaitGroup().Add(1)
		go bp.worker(i)
	}

	return nil
}

func (bp *BasePool[D]) Stop(stopTimeOut time.Duration) error {
	if !bp.running.CompareAndSwap(true, false) {
		return errs.NewCommonError(fmt.Sprintf("worker pool %s not running", bp.GetName()), nil)
	}

	bp.GetLogger().Debugf("worker pool %s stopping", bp.GetName())
	defer bp.GetLogger().Debugf("worker pool %s stopped", bp.GetName())

	// data channel
	close(bp.dataChan)

	// complete channel processing
	if !bp.GetConfig().CompleteProcess && bp.GetContextCancel() != nil {
		bp.GetLogger().Debugf("worker pool %s is not complete data channel processing, cancel pool context", bp.GetName())
		bp.GetContextCancel()()
	}

	// stop workers gracefully
	stopChan := make(chan struct{})
	go func() {
		bp.GetWaitGroup().Wait()
		close(stopChan)
	}()
	bp.GetLogger().Debugf("worker pool %s waiting for workers to stop", bp.GetName())
	select {
	case <-stopChan:
		bp.GetLogger().Debugf("worker pool %s stopped gracefully, all data processed", bp.GetName())
	case <-time.After(stopTimeOut):
		bp.GetLogger().Debugf("worker pool %s stop timed out, force stopping, some data not processed and will be lost", bp.GetName())
	}
	if bp.GetContextCancel() != nil {
		bp.GetContextCancel()()
	}

	return nil
}

func (bp *BasePool[D]) Push(data D) {
	if !bp.IsRunning() {
		return
	}

	select {
	case bp.dataChan <- data:
		bp.GetLogger().Debugf("worker pool %s push data [%v]", bp.GetName(), data)
	case <-bp.GetContext().Done():
		bp.GetLogger().Debugf("worker pool %s stop push by context", bp.GetName())
	}
}

func (bp *BasePool[D]) TryPush(data D) bool {
	if !bp.IsRunning() {
		return false
	}

	select {
	case bp.dataChan <- data:
		bp.GetLogger().Debugf("worker pool %s push data [%v]", bp.GetName(), data)

		return true
	case <-bp.GetContext().Done():
		bp.GetLogger().Debugf("worker pool %s stop push by context", bp.GetName())

		return false
	default:
		bp.GetLogger().Debugf("worker pool %s try push default, data [%v] ignored and lost", bp.GetName(), data)

		return false
	}
}

func (bp *BasePool[D]) Len() int {
	return len(bp.dataChan)
}

func (bp *BasePool[D]) Capacity() int {
	return cap(bp.dataChan)
}

func (bp *BasePool[D]) worker(workerIndex int) {
	bp.GetLogger().Debugf("worker pool %s worker %v start", bp.GetName(), workerIndex)
	defer bp.GetLogger().Debugf("worker pool %s worker %v finish", bp.GetName(), workerIndex)
	defer bp.GetWaitGroup().Done()

	for {
		select {
		case <-bp.GetContext().Done():
			bp.GetLogger().Debugf("worker pool %s worker %v context done, stop worker", bp.GetName(), workerIndex)
			return
		case data, opened := <-bp.dataChan:
			if !opened {
				bp.GetLogger().Debugf("worker pool %s worker %v queue closed, stop worker", bp.GetName(), workerIndex)
				return
			}
			if bp.jobHandler != nil {
				err := bp.jobHandler(bp.GetContext(), workerIndex, data)
				if err != nil {
					bp.GetLogger().Errorf("worker pool %s worker %v job failed:  %v", bp.GetName(), workerIndex, err)
				}
			} else {
				bp.GetLogger().Warnf("worker pool %s worker %v job handler not applied", bp.GetName(), workerIndex)
			}
		}
	}
}

func (bp *BasePool[D]) GetName() string {
	return bp.name
}

func (bp *BasePool[D]) GetContext() context.Context {
	return bp.ctx
}

func (bp *BasePool[D]) GetContextCancel() context.CancelFunc {
	return bp.cancel
}

func (bp *BasePool[D]) GetLogger() logger.Logger {
	return bp.log
}

func (bp *BasePool[D]) GetWaitGroup() *sync.WaitGroup {
	return &bp.wg
}

func (bp *BasePool[D]) GetConfig() *BasePoolConfig {
	return bp.config
}

func (bp *BasePool[D]) IsRunning() bool {
	return bp.running.Load()
}
