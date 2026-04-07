package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type BaseSchedulerDispatcherConfig struct {
	SchedulerConfig *BaseSchedulerConfig
	PoolConfig      *BasePoolConfig
}

func NewBaseSchedulerDispatcherConfig2(
	schedulerConfig *BaseSchedulerConfig,
	poolConfig *BasePoolConfig,
) *BaseSchedulerDispatcherConfig {
	return &BaseSchedulerDispatcherConfig{
		SchedulerConfig: schedulerConfig,
		PoolConfig:      poolConfig,
	}
}

type BaseSchedulerDispatcher[D comparable] struct {
	*BaseScheduler
	workerPool   Pool[D]
	config       *BaseSchedulerDispatcherConfig
	dataProvider DispatcherDataProvider[D]
}

var _ Scheduler = (*BaseSchedulerDispatcher[string])(nil)
var _ CommonWorker = (*BaseSchedulerDispatcher[string])(nil)

func NewBaseSchedulerDispatcher[D comparable](
	name string,
	config *BaseSchedulerDispatcherConfig,
	dataProvider DispatcherDataProvider[D],
	jobHandler JobHandler[D],
	log logger.Logger,
) *BaseSchedulerDispatcher[D] {
	res := &BaseSchedulerDispatcher[D]{
		config:       config,
		dataProvider: dataProvider,
		workerPool:   NewBasePool[D](name, config.PoolConfig, jobHandler, log),
	}

	// base
	res.BaseScheduler = NewBaseScheduler(name, res.timerDispatcher, config.SchedulerConfig, log)

	return res
}

func (bsd *BaseSchedulerDispatcher[D]) Start(ctx context.Context) error {
	errPool := bsd.workerPool.Start(ctx)
	errScheduler := bsd.BaseScheduler.Start(ctx)
	err := errors.Join(errPool, errScheduler)
	if err != nil {
		err = errors.Join(err, bsd.Stop(0))

		return errs.NewCommonError(fmt.Sprintf("scheduler dispatcher %s start failed", bsd.GetName()), err)
	}

	return nil
}

func (bsd *BaseSchedulerDispatcher[D]) Stop(stopTimeOut time.Duration) error {
	errScheduler := bsd.BaseScheduler.Stop(stopTimeOut)
	errPool := bsd.workerPool.Stop(stopTimeOut)
	err := errors.Join(errPool, errScheduler)
	if err != nil {
		return errs.NewCommonError(fmt.Sprintf("scheduler dispatcher %s stop failed", bsd.GetName()), err)
	}

	return nil
}

func (bsd *BaseSchedulerDispatcher[D]) timerDispatcher(eventTime time.Time) error {
	bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s start", bsd.GetName(), eventTime.Format(time.DateTime))
	defer bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s finish", bsd.GetName(), eventTime.Format(time.DateTime))

	if bsd.dataProvider == nil {
		return errs.NewCommonError(fmt.Sprintf("scheduler dispatcher %s time event %s data provider not applied", bsd.GetName(), eventTime.Format(time.DateTime)), nil)
	}

	res, err := bsd.dataProvider(bsd.GetContext(), eventTime)
	if err != nil {
		return err
	}
	bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s got %v data records", bsd.GetName(), eventTime.Format(time.DateTime), len(res))

	for _, data := range res {
		bsd.workerPool.Push(data)
	}

	return nil
}

func (bsd *BaseSchedulerDispatcher[D]) GetConfig() *BaseSchedulerDispatcherConfig {
	return bsd.config
}
