package worker

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
)

// Janitor представляет собой фоновый строго типизированный воркер-планировщик (Scheduler Worker).
//
// Инкапсулирует работу с BaseScheduler фреймворка, периодически запуская такт проактивной
// очистки просроченных по TTL записей (Garbage Collection) в оперативной памяти кэш-системы.
type Janitor[K comparable, V any] struct {
	*worker.BaseScheduler // Встраиваем базовый планировщик для управления жизненным циклом и тикерами
	opts                  *JanitorOptions[K, V]
}

// NewJanitor — фабричный конструктор фонового очистителя кэша.
// Прозрачно связывает метод CacheJanitor кэш-менеджера со встроенным циклом планировщика задач.
func NewJanitor[K comparable, V any](options ...JanitorOption[K, V]) (*Janitor[K, V], error) {
	opts := NewJanitorOptions[K, V]()
	for _, opt := range options {
		opt(opts)
	}
	if err := opts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewJanitor[K, V]", "janitor options validation failed", err)
	}

	// scheduler
	scheduler, err := worker.NewBaseScheduler(
		worker.WithSchedulerName(opts.Name),
		worker.WithSchedulerStartInterval(opts.StartInterval),
		worker.WithSchedulerScheduleInterval(opts.ScheduleInterval),
		worker.WithSchedulerStopTimeout(opts.StopTimeout),
		worker.WithSchedulerLogger(opts.Logger),
		worker.WithSchedulerTimerDispatcher(opts.Cache.CacheJanitor),
	)
	if err != nil {
		return nil, errs.NewTlCommonError("NewJanitor[K, V]", "base scheduler creation failed", err)
	}

	return &Janitor[K, V]{
		opts:          opts,
		BaseScheduler: scheduler,
	}, nil
}
