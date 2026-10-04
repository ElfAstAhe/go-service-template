package worker

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// JanitorOption defines a strongly-typed functional options signature designed to incrementally configure JanitorOptions instances.
type JanitorOption[K comparable, V any] func(options *JanitorOptions[K, V])

// JanitorOptions compositionally expands the baseline background scheduler parameters contract to support localized memory storage caches attributes.
type JanitorOptions[K comparable, V any] struct {
	*worker.BaseSchedulerOptions                   // Embedded framework-level foundational intervals configuration blueprint fields
	Cache                        cache.Cache[K, V] // Underlying data caching storage contract target managing proactive eviction tasks
}

// NewJanitorOptions acts as a generic structural factory constructor initializing and unpacking baseline scheduler intervals default values.
func NewJanitorOptions[K comparable, V any]() *JanitorOptions[K, V] {
	return &JanitorOptions[K, V]{
		BaseSchedulerOptions: worker.NewBaseSchedulerOptions(),
	}
}

// Validate triggers sequential verification checks cascading threshold validations from the baseline scheduler rules into the local properties.
func (jo *JanitorOptions[K, V]) Validate() error {
	if err := jo.BaseSchedulerOptions.Validate(); err != nil {
		return errs.NewTlCommonError("Validate", "scheduler options validation failed", err)
	}

	if utils.IsNil(jo.Cache) {
		return errs.NewTlCommonError("Validate", "cache required", nil)
	}

	return nil
}

// ====================================================================
// Fluent API для Janitor (Новые и Обертки)
// ====================================================================

// WithJanitorName populates flat descriptive names descriptors tags inside the nested core configuration model block.
func WithJanitorName[K comparable, V any](name string) JanitorOption[K, V] {
	return func(options *JanitorOptions[K, V]) {
		options.Name = name
	}
}

// WithJanitorStartInterval maps initialization cold delay sequences offsets attributes controls across underlying clock engines.
func WithJanitorStartInterval[K comparable, V any](startInterval time.Duration) JanitorOption[K, V] {
	return func(options *JanitorOptions[K, V]) {
		options.StartInterval = startInterval
	}
}

// WithJanitorScheduleInterval logs cyclic task recurrence gaps boundaries configurations metrics payload values parameters.
func WithJanitorScheduleInterval[K comparable, V any](scheduleInterval time.Duration) JanitorOption[K, V] {
	return func(options *JanitorOptions[K, V]) {
		options.ScheduleInterval = scheduleInterval
	}
}

// WithJanitorStopTimeout applies soft processing countdown execution barriers constraints parameters prior to triggering termination panics.
func WithJanitorStopTimeout[K comparable, V any](timeout time.Duration) JanitorOption[K, V] {
	return func(options *JanitorOptions[K, V]) {
		options.StopTimeout = timeout
	}
}

// WithJanitorLogger isolates specialized structural diagnostic telemetry reporting streams pipelines handles.
func WithJanitorLogger[K comparable, V any](log logger.Logger) JanitorOption[K, V] {
	return func(options *JanitorOptions[K, V]) {
		options.Logger = log
	}
}

// WithJanitorCache configures target application-level memory clusters layers managing reactive evictions procedures.
func WithJanitorCache[K comparable, V any](cache cache.Cache[K, V]) JanitorOption[K, V] {
	return func(options *JanitorOptions[K, V]) {
		options.Cache = cache
	}
}
