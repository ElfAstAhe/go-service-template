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
	defaultPoolWorkerCount     int           = 2
	defaultPoolDataCapacity    int           = 64
	defaultPoolCompleteProcess bool          = true
	defaultPoolStopTimeout     time.Duration = time.Second * 5
)

// BasePoolOption определяет функциональный тип для конфигурации опций (Fluent API).
type BasePoolOption[D any] func(*BasePoolOptions[D])

// BasePoolOptions содержит параметры рантайма, необходимые для безопасной сборки и работы worker pool.
type BasePoolOptions[D any] struct {
	// Name Наименование воркера
	Name string

	// WorkerCount Количество параллельно запущенных горутин-обработчиков
	WorkerCount int

	// DataCapacity Буферная емкость внутреннего канала задач (Backpressure window)
	DataCapacity int

	// CompleteProcess Флаг: вычитывать ли буфер до конца при закрытии канала (true) или тушить экстренно (false)
	CompleteProcess bool

	// StopTimeout Временной лимит (таймаут) на мягкое завершение обработки перед принудительным выходом
	StopTimeout time.Duration

	// Logger Логгер
	Logger logger.Logger

	// JobHandler обработчик
	JobHandler JobHandler[D]
}

// NewBasePoolOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewBasePoolOptions[D any]() *BasePoolOptions[D] {
	return &BasePoolOptions[D]{
		WorkerCount:     defaultPoolWorkerCount,
		DataCapacity:    defaultPoolDataCapacity,
		CompleteProcess: defaultPoolCompleteProcess,
		StopTimeout:     defaultPoolStopTimeout,
	}
}

// Validate проверяет корректность абсолютно всех опций рантайма перед сборкой worker pool.
func (bpo *BasePoolOptions[D]) Validate() error {
	if strings.TrimSpace(bpo.Name) == "" {
		return errs.NewTlCommonError("Validate", "name is required", nil)
	}
	if bpo.WorkerCount <= 0 {
		return errs.NewTlCommonError("Validate", "worker count is required", nil)
	}
	if bpo.DataCapacity <= 0 {
		return errs.NewTlCommonError("Validate", "data capacity is required", nil)
	}
	if bpo.StopTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "stop timeout is required", nil)
	}
	if utils.IsNil(bpo.JobHandler) {
		return errs.NewTlCommonError("Validate", "job handler is required", nil)
	}
	if utils.IsNil(bpo.Logger) {
		return errs.NewTlCommonError("Validate", "logger is required", nil)
	}

	return nil
}

// ====================================================================
// Fluent API методы для сборки опций worker pool
// ====================================================================

// WithPoolName настраивает наименование worker pool
func WithPoolName[D any](name string) BasePoolOption[D] {
	return func(options *BasePoolOptions[D]) {
		options.Name = name
	}
}

// WithPoolWorkerCount настраивает кол-во обработчиков
func WithPoolWorkerCount[D any](workerCount int) BasePoolOption[D] {
	return func(options *BasePoolOptions[D]) {
		options.WorkerCount = workerCount
	}
}

// WithPoolDataCapacity настраивает объём очереди данных обработки
func WithPoolDataCapacity[D any](dataCapacity int) BasePoolOption[D] {
	return func(options *BasePoolOptions[D]) {
		options.DataCapacity = dataCapacity
	}
}

// WithPoolCompleteProcess настраивает признак обработки всей очереди до заверешения/останова
func WithPoolCompleteProcess[D any](completeProcess bool) BasePoolOption[D] {
	return func(options *BasePoolOptions[D]) {
		options.CompleteProcess = completeProcess
	}
}

// WithPoolStopTimeout настраивает таймаут времени остановки worker pool
func WithPoolStopTimeout[D any](stopTimeout time.Duration) BasePoolOption[D] {
	return func(options *BasePoolOptions[D]) {
		options.StopTimeout = stopTimeout
	}
}

// WithPoolLogger настраивает worker pool logger
func WithPoolLogger[D any](logger logger.Logger) BasePoolOption[D] {
	return func(options *BasePoolOptions[D]) {
		options.Logger = logger
	}
}

// WithPoolJobHandler настраивает обработчик
func WithPoolJobHandler[D any](handler JobHandler[D]) BasePoolOption[D] {
	return func(options *BasePoolOptions[D]) {
		options.JobHandler = handler
	}
}
