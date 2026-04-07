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

type TimerDispatcher func(eventTime time.Time) error

type BaseSchedulerConfig struct {
	// schedule
	startInterval    time.Duration
	scheduleInterval time.Duration
}

func NewBaseSchedulerConfig(
	startInterval time.Duration,
	scheduleInterval time.Duration,
) *BaseSchedulerConfig {
	return &BaseSchedulerConfig{
		startInterval:    startInterval,
		scheduleInterval: scheduleInterval,
	}
}

type BaseScheduler struct {
	name string
	// context
	ctx    context.Context
	cancel context.CancelFunc
	// sync
	wg sync.WaitGroup
	// schedule
	timer           *time.Timer
	timerDispatcher TimerDispatcher
	// config
	config *BaseSchedulerConfig
	// logging
	log logger.Logger
	//
	running *atomic.Bool
}

var _ Scheduler = (*BaseScheduler)(nil)

func NewBaseScheduler(
	name string,
	timerDispatcher TimerDispatcher,
	config *BaseSchedulerConfig,
	log logger.Logger,
) *BaseScheduler {
	res := &BaseScheduler{
		name:            name,
		timerDispatcher: timerDispatcher,
		config:          config,
		log:             log.GetLogger(name),
		running:         new(atomic.Bool),
	}
	res.running.Store(false)

	return res
}

func (bs *BaseScheduler) Start(ctx context.Context) error {
	if !bs.running.CompareAndSwap(false, true) {
		return errs.NewCommonError(fmt.Sprintf("scheduler %s already started", bs.GetName()), nil)
	}

	bs.GetLogger().Debugf("scheduler %s starting", bs.GetName())
	defer bs.GetLogger().Debugf("scheduler %s started", bs.GetName())

	// context
	bs.ctx, bs.cancel = context.WithCancel(ctx)
	// timer
	if bs.timer == nil {
		bs.timer = time.NewTimer(bs.GetConfig().startInterval)
	} else {
		if !bs.timer.Stop() {
			select {
			case <-bs.timer.C:
			default:
			}
		}
		bs.timer.Reset(bs.GetConfig().startInterval)
	}
	// dispatcher
	bs.GetWaitGroup().Add(1)
	go bs.timerEventListener()

	return nil
}

func (bs *BaseScheduler) Stop(stopTimeOut time.Duration) error {
	if !bs.running.CompareAndSwap(true, false) {
		return errs.NewCommonError(fmt.Sprintf("scheduler %s is not running", bs.GetName()), nil)
	}

	// timer
	if bs.timer != nil {
		if !bs.timer.Stop() {
			select {
			case <-bs.timer.C:
			default:
			}
		}
	}
	// cancel ctx
	if bs.GetContextCancel() != nil {
		bs.GetContextCancel()()
	}

	// waiting for stop
	stopChan := make(chan struct{})
	go func() {
		bs.GetWaitGroup().Wait()
		close(stopChan)
	}()
	bs.GetLogger().Debugf("scheduler %s waiting for workers to stop", bs.GetName())
	select {
	case <-stopChan:
		bs.GetLogger().Debugf("scheduler %s stopped gracefully", bs.GetName())
	case <-time.After(stopTimeOut):
		bs.GetLogger().Debugf("scheduler %s stop timed out, force stopping", bs.GetName())
	}

	return nil
}

func (bs *BaseScheduler) timerEventListener() {
	bs.GetLogger().Debugf("scheduler %s timer event listener start", bs.GetName())
	defer bs.GetLogger().Debugf("scheduler %s timer event listener finish", bs.GetName())
	defer bs.GetWaitGroup().Done()

	for {
		select {
		case <-bs.GetContext().Done():
			bs.GetLogger().Debugf("scheduler %s context done, stop time event listener", bs.GetName())

			return
		case eventTime := <-bs.timer.C:
			bs.GetLogger().Debugf("scheduler %s timer event listener, time event fired: %s", bs.GetName(), eventTime.Format(time.DateTime))
			if bs.timerDispatcher != nil {
				if err := bs.timerDispatcher(eventTime); err != nil {
					bs.GetLogger().Errorf("scheduler %s time event %s dispatcher failed: %v", bs.GetName(), eventTime.Format(time.DateTime), err)
				}
			} else {
				bs.GetLogger().Warnf("scheduler %s time event %s dispatcher not applied", bs.GetName(), eventTime.Format(time.DateTime))
			}

			bs.timer.Reset(bs.GetConfig().scheduleInterval)
		}
	}
}

func (bs *BaseScheduler) GetName() string {
	return bs.name
}

func (bs *BaseScheduler) GetContext() context.Context {
	return bs.ctx
}

func (bs *BaseScheduler) GetContextCancel() context.CancelFunc {
	return bs.cancel
}

func (bs *BaseScheduler) GetWaitGroup() *sync.WaitGroup {
	return &bs.wg
}

func (bs *BaseScheduler) GetLogger() logger.Logger {
	return bs.log
}

func (bs *BaseScheduler) IsRunning() bool {
	return bs.running.Load()
}

func (bs *BaseScheduler) GetTimer() *time.Timer {
	return bs.timer
}

func (bs *BaseScheduler) GetConfig() *BaseSchedulerConfig {
	return bs.config
}
