package worker

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// Константы дефолтов для внутренней защиты рантайм-компонента (независимые от пакета config)
const (
	DefaultPeriodicDispatcherPoolWorkerCount           int           = 2
	DefaultPeriodicDispatcherPoolDataCapacity          int           = 64
	DefaultPeriodicDispatcherPoolCompleteProcess       bool          = true
	DefaultPeriodicDispatcherSchedulerStartInterval    time.Duration = time.Second
	DefaultPeriodicDispatcherSchedulerScheduleInterval time.Duration = time.Second * 5
	DefaultPeriodicDispatcherStopTimeout               time.Duration = time.Second * 5
)

// PeriodicDispatcherOption определяет функциональный тип для конфигурации опций (Fluent API).
type PeriodicDispatcherOption[D comparable] func(*PeriodicDispatcherOptions[D])

// PeriodicDispatcherOptions содержит параметры рантайма, необходимые для безопасной сборки и работы scheduler dispatcher.
type PeriodicDispatcherOptions[D comparable] struct {
	// Name Наименование
	Name string

	// WorkerCount Количество параллельно запущенных горутин-обработчиков
	WorkerCount int

	// DataCapacity Буферная емкость внутреннего канала задач (Backpressure window)
	DataCapacity int

	// CompleteProcess Флаг: вычитывать ли буфер до конца при закрытии канала (true) или тушить экстренно (false)
	CompleteProcess bool

	// StartInterval Первичная задержка (холодное смещение) перед самым первым тиком таймера
	StartInterval time.Duration

	// ScheduleInterval Фиксированный интервал периодического повторения задач (период)
	ScheduleInterval time.Duration

	// StopTimeout Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	StopTimeout time.Duration

	// Logger Логгер
	Logger logger.Logger

	// JobHandler обработчик worker pool
	JobHandler PoolJobHandler[D]

	// DataProvider поставщик данных
	DataProvider DispatcherDataProvider[D]
}

// NewPeriodicDispatcherOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewPeriodicDispatcherOptions[D comparable]() *PeriodicDispatcherOptions[D] {
	return &PeriodicDispatcherOptions[D]{
		WorkerCount:      DefaultPeriodicDispatcherPoolWorkerCount,
		DataCapacity:     DefaultPeriodicDispatcherPoolDataCapacity,
		CompleteProcess:  DefaultPeriodicDispatcherPoolCompleteProcess,
		StartInterval:    DefaultPeriodicDispatcherSchedulerStartInterval,
		ScheduleInterval: DefaultPeriodicDispatcherSchedulerScheduleInterval,
		StopTimeout:      DefaultPeriodicDispatcherStopTimeout,
	}
}

// Validate проверяет корректность абсолютно всех опций рантайма перед сборкой scheduler.
//
//goland:noinspection DuplicatedCode
func (sdo *PeriodicDispatcherOptions[D]) Validate() error {
	if strings.TrimSpace(sdo.Name) == "" {
		return errs.NewTlCommonError("Validate", "name is required", nil)
	}
	if sdo.WorkerCount <= 0 {
		return errs.NewTlCommonError("Validate", "worker count is required", nil)
	}
	if sdo.DataCapacity <= 0 {
		return errs.NewTlCommonError("Validate", "data capacity is required", nil)
	}
	if sdo.StopTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "stop timeout is required", nil)
	}
	if utils.IsNil(sdo.JobHandler) {
		return errs.NewTlCommonError("Validate", "job handler is required", nil)
	}
	if utils.IsNil(sdo.Logger) {
		return errs.NewTlCommonError("Validate", "logger is required", nil)
	}
	if utils.IsNil(sdo.DataProvider) {
		return errs.NewTlCommonError("Validate", "data provider is required", nil)
	}
	if sdo.StartInterval <= 0 {
		return errs.NewTlCommonError("Validate", "start interval is required", nil)
	}
	if sdo.ScheduleInterval <= 0 {
		return errs.NewTlCommonError("Validate", "schedule interval is required", nil)
	}

	return nil
}

// ====================================================================
// Fluent API методы для сборки опций scheduler dispatcher
// ====================================================================

// WithPeriodicDispatcherName настраивает наименование
func WithPeriodicDispatcherName[D comparable](name string) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.Name = name
	}
}

// WithPeriodicDispatcherPoolWorkerCount настраивает кол-во обработчиков
func WithPeriodicDispatcherPoolWorkerCount[D comparable](count int) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.WorkerCount = count
	}
}

// WithPeriodicDispatcherPoolDataCapacity настраивает объём очереди данных обработки
func WithPeriodicDispatcherPoolDataCapacity[D comparable](dataCapacity int) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.DataCapacity = dataCapacity
	}
}

// WithPeriodicDispatcherPoolCompleteProcess настраивает признак обработки всей очереди до заверешения/останова
func WithPeriodicDispatcherPoolCompleteProcess[D comparable](flag bool) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.CompleteProcess = flag
	}
}

// WithPeriodicDispatcherPoolJobHandler настраивает обработчик
func WithPeriodicDispatcherPoolJobHandler[D comparable](handler PoolJobHandler[D]) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.JobHandler = handler
	}
}

// WithPeriodicDispatcherSchedulerStartInterval настраивает первичную задержку (холодное смещение) перед самым первым тиком таймера
func WithPeriodicDispatcherSchedulerStartInterval[D comparable](startInterval time.Duration) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.StartInterval = startInterval
	}
}

// WithPeriodicDispatcherSchedulerScheduleInterval настраивает фиксированный интервал периодического повторения задач (период)
func WithPeriodicDispatcherSchedulerScheduleInterval[D comparable](scheduleInterval time.Duration) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.ScheduleInterval = scheduleInterval
	}
}

// WithPeriodicDispatcherStopTimeout настраивает временной лимит (таймаут) на мягкое завершение активной итерации обработчика
func WithPeriodicDispatcherStopTimeout[D comparable](timeout time.Duration) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.StopTimeout = timeout
	}
}

// WithPeriodicDispatcherDataProvider настраивает обработчик событияпостащика данных
func WithPeriodicDispatcherDataProvider[D comparable](dataProvider DispatcherDataProvider[D]) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.DataProvider = dataProvider
	}
}

// WithPeriodicDispatcherLogger настраивает scheduler dispatcher logger
func WithPeriodicDispatcherLogger[D comparable](logger logger.Logger) PeriodicDispatcherOption[D] {
	return func(options *PeriodicDispatcherOptions[D]) {
		options.Logger = logger
	}
}
