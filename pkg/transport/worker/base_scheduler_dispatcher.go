package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type JobHandler[D comparable] func(ctx context.Context, workerIndex int, data D) error
type DispatcherDataProvider[D comparable] func(ctx context.Context, eventTime time.Time) ([]D, error)

type BaseSchedulerDispatcherConfig struct {
	*BaseSchedulerConfig
	WorkerCount  int
	DataCapacity int
}

func NewBaseSchedulerDispatcherConfig(config *BaseSchedulerConfig, workerCount int, dataCapacity int) *BaseSchedulerDispatcherConfig {
	return &BaseSchedulerDispatcherConfig{
		BaseSchedulerConfig: config,
		WorkerCount:         workerCount,
		DataCapacity:        dataCapacity,
	}
}

type BaseSchedulerDispatcher[D comparable] struct {
	*BaseScheduler
	dataChan     chan D
	config       *BaseSchedulerDispatcherConfig
	dataProvider DispatcherDataProvider[D]
	jobHandler   JobHandler[D]
}

func NewBaseSchedulerDispatcher[D comparable](
	name string,
	parentCtx context.Context,
	config *BaseSchedulerDispatcherConfig,
	dataProvider DispatcherDataProvider[D],
	jobHandler JobHandler[D],
	log logger.Logger,
) *BaseSchedulerDispatcher[D] {
	res := &BaseSchedulerDispatcher[D]{
		config:       config,
		dataProvider: dataProvider,
		jobHandler:   jobHandler,
	}

	// base
	res.BaseScheduler = NewBaseScheduler(name, parentCtx, res.dispatch, config.BaseSchedulerConfig, log)

	return res
}

func (bsd *BaseSchedulerDispatcher[D]) Start() error {
	err := bsd.BaseScheduler.Start()
	if err != nil {
		return err
	}
	// data
	bsd.dataChan = make(chan D, bsd.config.DataCapacity)
	// workers
	for i := 0; i < bsd.config.WorkerCount; i++ {
		bsd.GetWaitGroup().Add(1)
		go bsd.worker(i)
	}

	return nil
}

func (bsd *BaseSchedulerDispatcher[D]) Stop() error {
	err := bsd.BaseScheduler.Stop()
	// data
	close(bsd.dataChan)

	return err
}

func (bsd *BaseSchedulerDispatcher[D]) dispatch(eventTime time.Time) error {
	bsd.GetLogger().Debugf("start %s dispatcher %s", bsd.GetName(), eventTime.Format(time.DateTime))
	defer bsd.GetLogger().Debugf("finish %s dispatcher %s", bsd.GetName(), eventTime.Format(time.DateTime))

	if bsd.dataProvider == nil {
		return errs.NewCommonError(fmt.Sprintf("failed %s dispatcher %s data provider not applied", bsd.GetName(), eventTime.Format(time.DateTime)), nil)
	}

	res, err := bsd.dataProvider(bsd.GetContext(), eventTime)
	if err != nil {
		return err
	}

	for _, data := range res {
		select {
		case bsd.dataChan <- data:
			bsd.GetLogger().Debugf("dispatcher %s event %s dispatch data [%v]", bsd.GetName(), eventTime.Format(time.DateTime), data)
		case <-bsd.GetContext().Done():
			bsd.GetLogger().Debugf("dispatcher %s event %s stop by context", bsd.GetName(), eventTime.Format(time.DateTime))
			return bsd.GetContext().Err()
		}
	}

	return nil
}

func (bsd *BaseSchedulerDispatcher[D]) worker(workerIndex int) {
	bsd.GetLogger().Debugf("start %s worker %v", bsd.GetName(), workerIndex)
	defer bsd.GetLogger().Debugf("finish %s worker %v", bsd.GetName(), workerIndex)
	defer bsd.GetWaitGroup().Done()

	for {
		select {
		case <-bsd.GetContext().Done():
			return
		case data, opened := <-bsd.dataChan:
			if !opened {
				bsd.GetLogger().Debugf("stop %s worker %v, queue closed", bsd.GetName(), workerIndex)
				return
			}
			if bsd.jobHandler != nil {
				err := bsd.jobHandler(bsd.GetContext(), workerIndex, data)
				if err != nil {
					bsd.GetLogger().Errorf("failed %s worker %v work job: %v", bsd.GetName(), workerIndex, err)
				}
			} else {
				bsd.GetLogger().Warnf("worker %s worker %v job handler not applied", bsd.GetName(), workerIndex)
			}
		}
	}
}

func (bsd *BaseSchedulerDispatcher[D]) GetConfig() *BaseSchedulerDispatcherConfig {
	return bsd.config
}
