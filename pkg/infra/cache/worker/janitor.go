package worker

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
)

// Janitor представляет собой фоновый строго типизированный воркер-планировщик (Scheduler Worker).
//
// Инкапсулирует работу с PeriodicScheduler фреймворка, периодически запуская такт проактивной
// очистки просроченных по TTL записей (Garbage Collection) в оперативной памяти кэш-системы.
type Janitor[K comparable, V any] struct {
	*worker.PeriodicScheduler // Встраиваем базовый планировщик для управления жизненным циклом и тикерами
	opts                      *JanitorOptions[K, V]
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
	scheduler, err := worker.NewPeriodicScheduler(
		worker.WithPeriodicSchedulerName(opts.Name),
		worker.WithPeriodicSchedulerStartInterval(opts.StartInterval),
		worker.WithPeriodicSchedulerScheduleInterval(opts.ScheduleInterval),
		worker.WithPeriodicSchedulerStopTimeout(opts.StopTimeout),
		worker.WithPeriodicSchedulerLogger(opts.Logger),
		worker.WithPeriodicSchedulerTimerDispatcher(opts.Cache.CacheJanitor),
	)
	if err != nil {
		return nil, errs.NewTlCommonError("NewJanitor[K, V]", "base scheduler creation failed", err)
	}

	return &Janitor[K, V]{
		opts:              opts,
		PeriodicScheduler: scheduler,
	}, nil
}
