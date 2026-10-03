package worker

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
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

	// DagaProvider обработчик scheduler
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

func (sdo *BaseSchedulerDispatcherOption[D]) Validate() error {

}
