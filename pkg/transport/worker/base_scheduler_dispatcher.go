package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

const schedulerDispatcherNameTemplate string = "scheduler-dispatcher-%s"

// BaseSchedulerDispatcher реализует интерфейсы CommonWorker, Scheduler и container.Runner,
// являясь строго типизированным оркестратором периодических задач (Generic Scheduler Dispatcher Pattern).
//
// Инкапсулирует под своим крылом базовый таймер (BaseScheduler) и внутренний многопоточный пул воркеров (Pool).
// За счет асинхронного веерного распределения задач полностью защищает системные таймеры Go от дрейфа времени,
// изолируя фазу извлечения данных (Data Fetching) от тяжелой параллельной бизнес-обработки (Job Processing).
type BaseSchedulerDispatcher[D comparable] struct {
	name         string
	ctx          context.Context
	cancel       context.CancelFunc
	scheduler    Scheduler                          // Встраивание (Embedding) базового планировщика для управления событиями таймера
	workerPool   Pool[D]                            // Ссылка на строго типизированный внутренний пул конкурентных воркеров
	dataProvider DispatcherDataProvider[D]          // Поставщик данных
	opts         *BaseSchedulerDispatcherOptions[D] // Указатель на агрегированную структуру конфигурационных параметров
	log          logger.Logger
}

// Гарантируем строгое соответствие контрактам и интерфейсам фреймворка на этапе компиляции
var _ CommonWorker = (*BaseSchedulerDispatcher[string])(nil)
var _ Scheduler = (*BaseSchedulerDispatcher[string])(nil)
var _ container.Runner = (*BaseSchedulerDispatcher[string])(nil)

// NewBaseSchedulerDispatcher — фабричный конструктор строго типизированного комплексного диспетчера.
// Автоматически инициализирует внутренний Pool[D], Scheduler и связывает его с циклом событий встроенного планировщика.
func NewBaseSchedulerDispatcher[D comparable](options ...BaseSchedulerDispatcherOption[D]) (*BaseSchedulerDispatcher[D], error) {
	opts := NewBaseSchedulerDispatcherOptions[D]()
	for _, opt := range options {
		opt(opts)
	}
	if err := opts.Validate(); err != nil {
		panic(err)
	}

	// Pool
	pool, err := NewBasePool[D](
		WithPoolName[D](opts.Name),
		WithPoolWorkerCount[D](opts.WorkerCount),
		WithPoolDataCapacity[D](opts.DataCapacity),
		WithPoolCompleteProcess[D](opts.CompleteProcess),
		WithPoolStopTimeout[D](opts.StopTimeout),
		WithPoolLogger[D](opts.Logger),
		WithPoolJobHandler[D](opts.JobHandler),
	)
	if err != nil {
		return nil, errs.NewTlCommonError("NewBaseSchedulerDispatcher", "pool creation failed", err)
	}

	// instance
	res := &BaseSchedulerDispatcher[D]{
		name:         fmt.Sprintf(schedulerDispatcherNameTemplate, opts.Name),
		workerPool:   pool,
		opts:         opts,
		log:          opts.Logger.GetLogger(fmt.Sprintf(schedulerDispatcherNameTemplate, opts.Name)),
		dataProvider: opts.DataProvider,
	}

	// scheduler
	scheduler, err := NewBaseScheduler(
		WithSchedulerName(opts.Name),
		WithSchedulerStartInterval(opts.StartInterval),
		WithSchedulerScheduleInterval(opts.ScheduleInterval),
		WithSchedulerStopTimeout(opts.StopTimeout),
		WithSchedulerLogger(opts.Logger),
		WithSchedulerTimerDispatcher(res.timerDispatcher),
	)
	if err != nil {
		return nil, errs.NewTlCommonError("NewBaseSchedulerDispatcher", "scheduler creation failed", err)
	}

	// Настраиваем instance
	res.scheduler = scheduler

	return res, nil
}

// Start выполняет каскадный запуск рантайма: сначала переводит в активное состояние внутренний пул воркеров,
// а затем взводит циклический таймер планировщика.
// В случае сбоя старта любого из компонентов инициирует экстренный автоматический откат (Stop)
// в рамках жесткого лимита времени (100мс) и возвращает агрегированную системную ошибку CommonError.
func (bsd *BaseSchedulerDispatcher[D]) Start(ctx context.Context) error {
	bsd.GetLogger().Debugf("scheduler dispatcher %s starting", bsd.GetName())
	defer bsd.GetLogger().Debugf("scheduler dispatcher %s started", bsd.GetName())

	bsd.ctx, bsd.cancel = context.WithCancel(ctx)

	if err := bsd.startWorkers(bsd.ctx, bsd.workerPool, bsd.scheduler); err != nil {
		// Фаза автоматического Fail-Fast отката: принудительно гасим частично поднятые ресурсы
		stopCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err = errors.Join(err, bsd.Stop(stopCtx))

		return errs.NewTlCommonError("Start", fmt.Sprintf("scheduler dispatcher %s start failed", bsd.GetName()), err)
	}

	return nil
}

// Stop осуществляет синхронную мягкую остановку (Graceful Shutdown) всей экосистемы диспетчера.
// Сначала деактивирует и очищает системный таймер в ядре ОС, блокируя генерацию новых событий,
// а затем каскадно останавливает внутренний пул воркеров, давая им доварить текущие задачи.
// Агрегирует ошибки останова обоих узлов через errors.Join, возвращая CommonError при сбоях.
func (bsd *BaseSchedulerDispatcher[D]) Stop(stopCtx context.Context) error {
	errScheduler := bsd.scheduler.Stop(stopCtx)
	errPool := bsd.workerPool.Stop(stopCtx)
	err := errors.Join(errPool, errScheduler)
	if err != nil {
		return errs.NewCommonError(fmt.Sprintf("scheduler dispatcher %s stop failed", bsd.GetName()), err)
	}

	return nil
}

// startWorkers performs concurrent startup routines across all registered core workers engines aggregates.
func (bsd *BaseSchedulerDispatcher[D]) startWorkers(ctx context.Context, workers ...CommonWorker) error {
	var startWG sync.WaitGroup
	errList := utils.NewConcurrentList[error]()

	for _, localWorker := range workers {
		bsd.log.Debugf("worker %s starting", localWorker.GetName())
		startWG.Add(1)
		go func() {
			defer startWG.Done()

			if err := localWorker.Start(ctx); err != nil {
				errList.Append(err)
			}
		}()
	}

	startWG.Wait()

	if err := errors.Join(errList.Snapshot()...); err != nil {
		return errs.NewTlCommonError("startWorkers", "scheduler dispatcher start workers failed", err)
	}

	return nil
}

// stopWorkers executes simultaneous teardown commands pacing multiple background routines processing elements.
func (bsd *BaseSchedulerDispatcher[D]) stopWorkers(ctx context.Context, workers ...CommonWorker) error {
	var stopWG sync.WaitGroup
	errList := utils.NewConcurrentList[error]()
	for _, localWorker := range workers {
		bsd.log.Debugf("worker %s stopping", localWorker.GetName())
		stopWG.Add(1)
		go func() {
			defer stopWG.Done()
			if err := localWorker.Stop(ctx); err != nil {
				errList.Append(err)
			}
		}()
	}
	stopWG.Wait()
	if err := errors.Join(errList.Snapshot()...); err != nil {
		return errs.NewTlCommonError("stopWorkers", "scheduler dispatcher stop workers failed", err)
	}

	return nil
}

// timerDispatcher — внутренний (неэкспортируемый) мост-колбек, связывающий тики таймера с очередью пула.
// Вызывается базовым планировщиком, извлекает пачку данных через dataProvider и веерно распределяет их по воркерам.
func (bsd *BaseSchedulerDispatcher[D]) timerDispatcher(ctx context.Context, eventTime time.Time) error {
	bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s start", bsd.GetName(), eventTime.Format(time.DateTime))
	defer bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s finish", bsd.GetName(), eventTime.Format(time.DateTime))

	if bsd.dataProvider == nil {
		return errs.NewCommonError(fmt.Sprintf("scheduler dispatcher %s time event %s data provider not applied", bsd.GetName(), eventTime.Format(time.DateTime)), nil)
	}

	// Извлекаем массив задач из внешнего источника (БД, кэш, API) средствами прикладного провайдера ограничение по времени на стороне потребителя компоненты
	res, err := bsd.dataProvider(bsd.GetContext(), eventTime)
	if err != nil {
		return err
	}
	bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s got %v data records", bsd.GetName(), eventTime.Format(time.DateTime), len(res))

	// Итерируемся по пачке извлеченных данных для их маршутизации в очередь воркеров
	for _, data := range res {
		select {
		case <-ctx.Done():
			// 🛡️ Защитный барьер: если контекст отменен (SIGTERM), немедленно прекращаем раздачу задач пулу
			bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s break by context done", bsd.GetName(), eventTime.Format(time.DateTime))

			return ctx.Err()
		default:
			// Блокирующий или неблокирующий (зависит от настроек канала) сброс задачи в буфер пула воркеров
			bsd.workerPool.Push(data)
		}
	}

	return nil
}

// GetName resolves active literal instance labels.
func (bsd *BaseSchedulerDispatcher[D]) GetName() string {
	return bsd.name
}

// GetContext retrieves active generic scheduler loop lifecycle contexts.
func (bsd *BaseSchedulerDispatcher[D]) GetContext() context.Context {
	return bsd.ctx
}

// GetContextCancel extracts context termination closure delegates maps hooks.
func (bsd *BaseSchedulerDispatcher[D]) GetContextCancel() context.CancelFunc {
	return bsd.cancel
}

// GetLogger provides component structured reporting diagnostic handles.
func (bsd *BaseSchedulerDispatcher[D]) GetLogger() logger.Logger {
	return bsd.log
}

// GetScheduler exposes embedded base low-level core clock structures triggers.
func (bsd *BaseSchedulerDispatcher[D]) GetScheduler() Scheduler {
	return bsd.scheduler
}

// GetWorkerPool exposes nested strongly-typed concurrency pool handles execution structures.
func (bsd *BaseSchedulerDispatcher[D]) GetWorkerPool() Pool[D] {
	return bsd.workerPool
}

// IsRunning evaluates operational pipeline active thresholds.
func (bsd *BaseSchedulerDispatcher[D]) IsRunning() bool {
	return bsd.workerPool.IsRunning() && bsd.scheduler.IsRunning()
}

// GetWaitGroup targets internal sync tracking aggregates arrays layouts blueprints.
func (bsd *BaseSchedulerDispatcher[D]) GetWaitGroup() *sync.WaitGroup {
	return nil
}

// GetOpts возвращает ссылку на конфигурационный паспорт агрегированных
func (bsd *BaseSchedulerDispatcher[D]) GetOpts() *BaseSchedulerDispatcherOptions[D] {
	return bsd.opts
}
