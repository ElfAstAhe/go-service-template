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
	defaultSchedulerDispatcherPoolWorkerCount           int           = 2
	defaultSchedulerDispatcherPoolDataCapacity          int           = 64
	defaultSchedulerDispatcherPoolCompleteProcess       bool          = true
	defaultSchedulerDispatcherSchedulerStartInterval    time.Duration = time.Second
	defaultSchedulerDispatcherSchedulerScheduleInterval time.Duration = time.Second * 5
	defaultSchedulerDispatcherStopTimeout               time.Duration = time.Second * 5
)

// BaseSchedulerDispatcherOption определяет функциональный тип для конфигурации опций (Fluent API).
type BaseSchedulerDispatcherOption[D comparable] func(*BaseSchedulerDispatcherOptions[D])

// BaseSchedulerDispatcherOptions содержит параметры рантайма, необходимые для безопасной сборки и работы scheduler dispatcher.
type BaseSchedulerDispatcherOptions[D comparable] struct {
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
	JobHandler JobHandler[D]

	// DataProvider поставщик данных
	DataProvider DispatcherDataProvider[D]
}

// NewBaseSchedulerDispatcherOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewBaseSchedulerDispatcherOptions[D comparable]() *BaseSchedulerDispatcherOptions[D] {
	return &BaseSchedulerDispatcherOptions[D]{
		WorkerCount:      defaultSchedulerDispatcherPoolWorkerCount,
		DataCapacity:     defaultSchedulerDispatcherPoolDataCapacity,
		CompleteProcess:  defaultSchedulerDispatcherPoolCompleteProcess,
		StartInterval:    defaultSchedulerDispatcherSchedulerStartInterval,
		ScheduleInterval: defaultSchedulerDispatcherSchedulerScheduleInterval,
		StopTimeout:      defaultSchedulerDispatcherStopTimeout,
	}
}

// Validate проверяет корректность абсолютно всех опций рантайма перед сборкой scheduler.
//
//goland:noinspection DuplicatedCode
func (sdo *BaseSchedulerDispatcherOptions[D]) Validate() error {
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

// WithSchedulerDispatcherName настраивает наименование
func WithSchedulerDispatcherName[D comparable](name string) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.Name = name
	}
}

// WithSchedulerDispatcherPoolWorkerCount настраивает кол-во обработчиков
func WithSchedulerDispatcherPoolWorkerCount[D comparable](count int) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.WorkerCount = count
	}
}

// WithSchedulerDispatcherPoolDataCapacity настраивает объём очереди данных обработки
func WithSchedulerDispatcherPoolDataCapacity[D comparable](dataCapacity int) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.DataCapacity = dataCapacity
	}
}

// WithSchedulerDispatcherPoolCompleteProcess настраивает признак обработки всей очереди до заверешения/останова
func WithSchedulerDispatcherPoolCompleteProcess[D comparable](flag bool) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.CompleteProcess = flag
	}
}

// WithSchedulerDispatcherPoolJobHandler настраивает обработчик
func WithSchedulerDispatcherPoolJobHandler[D comparable](handler JobHandler[D]) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.JobHandler = handler
	}
}

// WithSchedulerDispatcherSchedulerStartInterval настраивает первичную задержку (холодное смещение) перед самым первым тиком таймера
func WithSchedulerDispatcherSchedulerStartInterval[D comparable](startInterval time.Duration) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.StartInterval = startInterval
	}
}

// WithSchedulerDispatcherSchedulerScheduleInterval настраивает фиксированный интервал периодического повторения задач (период)
func WithSchedulerDispatcherSchedulerScheduleInterval[D comparable](scheduleInterval time.Duration) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.ScheduleInterval = scheduleInterval
	}
}

// WithSchedulerDispatcherStopTimeout настраивает временной лимит (таймаут) на мягкое завершение активной итерации обработчика
func WithSchedulerDispatcherStopTimeout[D comparable](timeout time.Duration) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.StopTimeout = timeout
	}
}

// WithSchedulerDispatcherDataProvider настраивает обработчик событияпостащика данных
func WithSchedulerDispatcherDataProvider[D comparable](dataProvider DispatcherDataProvider[D]) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.DataProvider = dataProvider
	}
}

// WithSchedulerDispatcherLogger настраивает scheduler dispatcher logger
func WithSchedulerDispatcherLogger[D comparable](logger logger.Logger) BaseSchedulerDispatcherOption[D] {
	return func(options *BaseSchedulerDispatcherOptions[D]) {
		options.Logger = logger
	}
}
