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

const periodicSchedulerNameTemplate string = "periodic-scheduler-%s"

// PeriodicScheduler реализует интерфейсы Common, Scheduler и container.Runner,
// представляя собой отказоустойчивую базовую основу фонового планировщика задач (Cron/Scheduler Base).
//
// Управляет циклом выполнения периодических процедур (например, Janitor-воркеров очистки сессий)
// на базе низкоуровневых таймеров ядра Go, гарантируя отсутствие дрейфа времени и мягкий Graceful Shutdown.
type PeriodicScheduler struct {
	name            string
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	timer           *time.Timer
	timerDispatcher TimerDispatcher
	opts            *PeriodicSchedulerOptions
	log             logger.Logger
	running         *atomic.Bool
}

// Гарантируем соответствие контрактам и интерфейсам на этапе компиляции
var _ Common = (*PeriodicScheduler)(nil)
var _ Scheduler = (*PeriodicScheduler)(nil)
var _ container.Runner = (*PeriodicScheduler)(nil)

// NewPeriodicScheduler — фабричный конструктор базового планировщика периодических задач.
func NewPeriodicScheduler(options ...PeriodicSchedulerOption) (*PeriodicScheduler, error) {
	opts := NewPeriodicSchedulerOptions()
	for _, option := range options {
		option(opts)
	}
	if err := opts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewPeriodicScheduler", "validate option failed", err)
	}

	res := &PeriodicScheduler{
		name:            fmt.Sprintf(periodicSchedulerNameTemplate, opts.Name),
		timerDispatcher: opts.TimerDispatcher,
		opts:            opts,
		log:             opts.Logger.GetLogger(fmt.Sprintf(periodicSchedulerNameTemplate, opts.Name)),
		running:         new(atomic.Bool),
	}
	res.running.Store(false)

	return res, nil
}

// Start осуществляет атомарный запуск цикла планировщика и взводит таймер на стартовое смещение StartInterval.
// Выделяет внутренний контекст выполнения и разворачивает асинхронную горутину прослушивания тиков.
// Возвращает CommonError, если планировщик уже запущен (защита от дублирования потоков выполнения).
func (bs *PeriodicScheduler) Start(ctx context.Context) error {
	if !bs.running.CompareAndSwap(false, true) {
		return nil
	}

	bs.GetLogger().Debugf("periodic scheduler %s Start start", bs.GetName())
	defer bs.GetLogger().Debugf("periodic scheduler %s Start finish", bs.GetName())

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
	bs.wg.Add(1)
	go bs.timerEventListener()

	return nil
}

// Stop осуществляет мягкое контролируемое гашение планировщика (Graceful Shutdown).
// Принудительно очищает и деактивирует системные таймеры в ядре ОС, отменяет контекст выполнения
// фонового Event Loop и ожидает физического завершения активной итерации обработчика в рамках лимита StopTimeout.
// Возвращает CommonError, если компонент не находится в состоянии выполнения.
func (bs *PeriodicScheduler) Stop(stopCtx context.Context) error {
	if !bs.running.CompareAndSwap(true, false) {
		return nil
	}

	if bs.timer != nil {
		if !bs.timer.Stop() {
			select {
			case <-bs.timer.C:
			default:
			}
		}
	}
	if bs.cancel != nil {
		bs.cancel()
	}

	stopChan := make(chan struct{})
	go func() {
		bs.wg.Wait()
		close(stopChan)
	}()
	bs.GetLogger().Debugf("periodic scheduler %s waiting for workers to stop", bs.GetName())
	select {
	case <-stopChan:
		bs.GetLogger().Debugf("periodic scheduler %s stopped gracefully", bs.GetName())
	case <-time.After(bs.opts.StopTimeout):
		bs.GetLogger().Debugf("periodic scheduler %s stop timed out, force stopping", bs.GetName())
	case <-stopCtx.Done():
		bs.GetLogger().Debugf("periodic scheduler %s stopped by stop context, force stopping", bs.GetName())
	}

	return nil
}

// GetName возвращает уникальное текстовое наименование текущего экземпляра планировщика.
func (bs *PeriodicScheduler) GetName() string {
	return bs.name
}

// IsRunning возвращает текущий атомарный статус активности и здоровья планировщика (true — запущен и слушает тики).
func (bs *PeriodicScheduler) IsRunning() bool {
	return bs.running.Load()
}

// timerEventListener — внутренний (неэкспортируемый) бесконечный цикл обработки временных тиков (Event Loop).
func (bs *PeriodicScheduler) timerEventListener() {
	bs.GetLogger().Debugf("periodic scheduler %s timeEventListener start", bs.GetName())
	defer bs.GetLogger().Debugf("periodic scheduler %s timerEventListener finish", bs.GetName())
	defer bs.wg.Done()

	for {
		select {
		case <-bs.GetContext().Done():
			bs.GetLogger().Debugf("periodic scheduler %s context done, stop time event listener", bs.GetName())

			return
		case eventTime := <-bs.timer.C:
			bs.GetLogger().Debugf("periodic scheduler %s timer event listener, time event fired: %s", bs.GetName(), eventTime.Format(time.DateTime))
			if bs.timerDispatcher != nil {
				if err := bs.timerDispatcher(bs.ctx, eventTime); err != nil {
					bs.GetLogger().Errorf("periodic scheduler %s time event %s dispatcher failed: %v", bs.GetName(), eventTime.Format(time.DateTime), err)
				}
			} else {
				bs.GetLogger().Warnf("periodic scheduler %s time event %s dispatcher not applied", bs.GetName(), eventTime.Format(time.DateTime))
			}

			bs.timer.Reset(bs.GetOpts().ScheduleInterval)
		}
	}
}

// GetContext возвращает ссылку на внутренний контекст context.Context времени жизни планировщика.
func (bs *PeriodicScheduler) GetContext() context.Context {
	return bs.ctx
}

// GetLogger возвращает изолированный инстанс структурированного логгера, закрепленный за планировщиком.
func (bs *PeriodicScheduler) GetLogger() logger.Logger {
	return bs.log
}

// GetTimer возвращает указатель на нативный системный объект низкоуровневого таймера *time.Timer ядра Go.
func (bs *PeriodicScheduler) GetTimer() *time.Timer {
	return bs.timer
}

// GetOpts возвращает ссылку на конфигурационный паспорт параметров интервалов планировщика PeriodicSchedulerOptions.
func (bs *PeriodicScheduler) GetOpts() *PeriodicSchedulerOptions {
	return bs.opts
}
