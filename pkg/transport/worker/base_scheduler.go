package worker

import (
	"context"
	"sync"
	"time"

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
	parent context.Context
	name   string
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
}

var _ Scheduler = (*BaseScheduler)(nil)

func NewBaseScheduler(
	name string,
	parent context.Context,
	timerDispatcher TimerDispatcher,
	config *BaseSchedulerConfig,
	log logger.Logger,
) *BaseScheduler {
	return &BaseScheduler{
		parent:          parent,
		name:            name,
		timerDispatcher: timerDispatcher,
		config:          config,
		log:             log.GetLogger(name),
	}
}

func (bs *BaseScheduler) Start() error {
	bs.log.Debugf("starting %s scheduler dispatcher", bs.name)
	defer bs.log.Debugf("started %s scheduler dispatcher", bs.name)

	// context
	bs.ctx, bs.cancel = context.WithCancel(bs.parent)
	// timer
	if bs.timer == nil {
		bs.timer = time.NewTimer(bs.config.startInterval)
	} else {
		if !bs.timer.Stop() {
			select {
			case <-bs.timer.C:
			default:
			}
		}
		bs.timer.Reset(bs.config.startInterval)
	}
	// dispatcher
	bs.wg.Add(1)
	go bs.timerEventListener()

	return nil
}

func (bs *BaseScheduler) Stop() error {
	// timer
	if bs.timer != nil {
		bs.timer.Stop()
	}
	// cancel ctx
	if bs.cancel != nil {
		bs.cancel()
	}

	// waiting for stop
	bs.wg.Wait()

	return nil
}

func (bs *BaseScheduler) timerEventListener() {
	bs.log.Debugf("start %s timer event listener", bs.name)
	defer bs.log.Debugf("finish %s timer event listener", bs.name)
	defer bs.wg.Done()

	for {
		select {
		case <-bs.ctx.Done():
			bs.log.Debugf("stop %s timer event listener by context", bs.name)

			return
		case eventTime := <-bs.timer.C:
			if bs.timerDispatcher != nil {
				if err := bs.timerDispatcher(eventTime); err != nil {
					bs.log.Errorf("dispatcher %s of timer event listener failed", bs.name)
				}
			}
			bs.timer.Reset(bs.config.scheduleInterval)
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

func (bs *BaseScheduler) GetTimer() *time.Timer {
	return bs.timer
}

func (bs *BaseScheduler) GetLogger() logger.Logger {
	return bs.log
}

func (bs *BaseScheduler) GetSchedulerConfig() *BaseSchedulerConfig {
	return bs.config
}

func (bs *BaseScheduler) GetWaitGroup() *sync.WaitGroup {
	return &bs.wg
}

func (bs *BaseScheduler) GetConfig() *BaseSchedulerConfig {
	return bs.config
}
