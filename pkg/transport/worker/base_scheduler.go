package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// BaseScheduler реализует интерфейсы CommonWorker, Scheduler и container.Runner,
// представляя собой отказоустойчивую базовую основу фонового планировщика задач (Cron/Scheduler Base).
//
// Управляет циклом выполнения периодических процедур (например, Janitor-воркеров очистки сессий)
// на базе низкоуровневых таймеров ядра Go, гарантируя отсутствие дрейфа времени и мягкий Graceful Shutdown.
type BaseScheduler struct {
	name            string
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	timer           *time.Timer
	timerDispatcher TimerDispatcher
	opts            *BaseSchedulerOptions
	log             logger.Logger
	running         *atomic.Bool
}

// Гарантируем соответствие контрактам и интерфейсам на этапе компиляции
var _ CommonWorker = (*BaseScheduler)(nil)
var _ Scheduler = (*BaseScheduler)(nil)
var _ container.Runner = (*BaseScheduler)(nil)

// NewBaseScheduler — фабричный конструктор базового планировщика периодических задач.
func NewBaseScheduler(options ...BaseSchedulerOption) (*BaseScheduler, error) {
	opts := NewBaseSchedulerOptions()
	for _, option := range options {
		option(opts)
	}
	if err := opts.Validate(); err != nil {

	}

	res := &BaseScheduler{
		name:            opts.Name,
		timerDispatcher: opts.TimerDispatcher,
		opts:            opts,
		log:             opts.Logger.GetLogger(opts.Name),
		running:         new(atomic.Bool),
	}
	res.running.Store(false)

	return res, nil
}

// Start осуществляет атомарный запуск цикла планировщика и взводит таймер на стартовое смещение StartInterval.
// Выделяет внутренний контекст выполнения и разворачивает асинхронную горутину прослушивания тиков.
// Возвращает CommonError, если планировщик уже запущен (защита от дублирования потоков выполнения).
func (bs *BaseScheduler) Start(ctx context.Context) error {
	if !bs.running.CompareAndSwap(false, true) {
		return errs.NewCommonError(fmt.Sprintf("scheduler %s already started", bs.GetName()), nil)
	}

	bs.GetLogger().Debugf("scheduler %s starting", bs.GetName())
	defer bs.GetLogger().Debugf("scheduler %s started", bs.GetName())

	//nolint:gosec // G118 : worker background context
	bs.ctx, bs.cancel = context.WithCancel(ctx)

	if bs.timer == nil {
		bs.timer = time.NewTimer(bs.GetOpts().StartInterval)
	} else {
		if !bs.timer.Stop() {
			select {
			case <-bs.timer.C:
			default:
			}
		}
		bs.timer.Reset(bs.GetOpts().StartInterval)
	}
	bs.GetWaitGroup().Add(1)
	go bs.timerEventListener()

	return nil
}

// Stop осуществляет мягкое контролируемое гашение планировщика (Graceful Shutdown).
// Принудительно очищает и деактивирует системные таймеры в ядре ОС, отменяет контекст выполнения
// фонового Event Loop и ожидает физического завершения активной итерации обработчика в рамках лимита StopTimeout.
// Возвращает CommonError, если компонент не находится в состоянии выполнения.
func (bs *BaseScheduler) Stop(stopCtx context.Context) error {
	if !bs.running.CompareAndSwap(true, false) {
		return errs.NewCommonError(fmt.Sprintf("scheduler %s is not running", bs.GetName()), nil)
	}

	if bs.timer != nil {
		if !bs.timer.Stop() {
			select {
			case <-bs.timer.C:
			default:
			}
		}
	}
	if bs.GetContextCancel() != nil {
		bs.GetContextCancel()()
	}

	stopChan := make(chan struct{})
	go func() {
		bs.GetWaitGroup().Wait()
		close(stopChan)
	}()
	bs.GetLogger().Debugf("scheduler %s waiting for workers to stop", bs.GetName())
	select {
	case <-stopChan:
		bs.GetLogger().Debugf("scheduler %s stopped gracefully", bs.GetName())
	case <-time.After(bs.opts.StopTimeout):
		bs.GetLogger().Debugf("scheduler %s stop timed out, force stopping", bs.GetName())
	case <-stopCtx.Done():
		bs.GetLogger().Debugf("scheduler %s stopped by stop context, force stopping", bs.GetName())
	}

	return nil
}

// timerEventListener — внутренний (неэкспортируемый) бесконечный цикл обработки временных тиков (Event Loop).
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
				if err := bs.timerDispatcher(bs.ctx, eventTime); err != nil {
					bs.GetLogger().Errorf("scheduler %s time event %s dispatcher failed: %v", bs.GetName(), eventTime.Format(time.DateTime), err)
				}
			} else {
				bs.GetLogger().Warnf("scheduler %s time event %s dispatcher not applied", bs.GetName(), eventTime.Format(time.DateTime))
			}

			bs.timer.Reset(bs.GetOpts().ScheduleInterval)
		}
	}
}

// GetName возвращает уникальное текстовое наименование текущего экземпляра планировщика.
func (bs *BaseScheduler) GetName() string {
	return bs.name
}

// GetContext возвращает ссылку на внутренний контекст context.Context времени жизни планировщика.
func (bs *BaseScheduler) GetContext() context.Context {
	return bs.ctx
}

// GetContextCancel возвращает функцию-триггер context.CancelFunc для принудительной плановой или экстренной отмены контекста.
func (bs *BaseScheduler) GetContextCancel() context.CancelFunc {
	return bs.cancel
}

// GetWaitGroup возвращает указатель на общую структуру sync.WaitGroup контроля запущенных дочерних потоков.
func (bs *BaseScheduler) GetWaitGroup() *sync.WaitGroup {
	return &bs.wg
}

// GetLogger возвращает изолированный инстанс структурированного логгера, закрепленный за планировщиком.
func (bs *BaseScheduler) GetLogger() logger.Logger {
	return bs.log
}

// IsRunning возвращает текущий атомарный статус активности и здоровья планировщика (true — запущен и слушает тики).
func (bs *BaseScheduler) IsRunning() bool {
	return bs.running.Load()
}

// GetTimer возвращает указатель на нативный системный объект низкоуровневого таймера *time.Timer ядра Go.
func (bs *BaseScheduler) GetTimer() *time.Timer {
	return bs.timer
}

// GetOpts возвращает ссылку на конфигурационный паспорт параметров интервалов планировщика BaseSchedulerOptions.
func (bs *BaseScheduler) GetOpts() *BaseSchedulerOptions {
	return bs.opts
}
