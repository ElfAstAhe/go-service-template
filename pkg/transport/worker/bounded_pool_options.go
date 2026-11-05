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
	DefaultBoundedPoolWorkerCount     int  = 2
	DefaultBoundedPoolDataCapacity    int  = 64
	DefaultBoundedPoolCompleteProcess bool = true
	DefaultBoundedPoolStopTimeout          = time.Second * 5
)

// BoundedPoolOption определяет функциональный тип для конфигурации опций (Fluent API).
type BoundedPoolOption[D any] func(*BoundedPoolOptions[D])

// BoundedPoolOptions содержит параметры рантайма, необходимые для безопасной сборки и работы worker pool.
type BoundedPoolOptions[D any] struct {
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
	JobHandler PoolJobHandler[D]
}

// NewBoundedPoolOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewBoundedPoolOptions[D any]() *BoundedPoolOptions[D] {
	return &BoundedPoolOptions[D]{
		WorkerCount:     DefaultBoundedPoolWorkerCount,
		DataCapacity:    DefaultBoundedPoolDataCapacity,
		CompleteProcess: DefaultBoundedPoolCompleteProcess,
		StopTimeout:     DefaultBoundedPoolStopTimeout,
	}
}

// Validate проверяет корректность абсолютно всех опций рантайма перед сборкой worker pool.
//
//goland:noinspection DuplicatedCode
func (bpo *BoundedPoolOptions[D]) Validate() error {
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

// WithBoundedPoolName настраивает наименование worker pool
func WithBoundedPoolName[D any](name string) BoundedPoolOption[D] {
	return func(options *BoundedPoolOptions[D]) {
		options.Name = name
	}
}

// WithBoundedPoolWorkerCount настраивает кол-во обработчиков
func WithBoundedPoolWorkerCount[D any](workerCount int) BoundedPoolOption[D] {
	return func(options *BoundedPoolOptions[D]) {
		options.WorkerCount = workerCount
	}
}

// WithBoundedPoolDataCapacity настраивает объём очереди данных обработки
func WithBoundedPoolDataCapacity[D any](dataCapacity int) BoundedPoolOption[D] {
	return func(options *BoundedPoolOptions[D]) {
		options.DataCapacity = dataCapacity
	}
}

// WithPoolCompleteProcess настраивает признак обработки всей очереди до заверешения/останова
func WithPoolCompleteProcess[D any](completeProcess bool) BoundedPoolOption[D] {
	return func(options *BoundedPoolOptions[D]) {
		options.CompleteProcess = completeProcess
	}
}

// WithBoundedPoolStopTimeout настраивает таймаут времени остановки worker pool
func WithBoundedPoolStopTimeout[D any](stopTimeout time.Duration) BoundedPoolOption[D] {
	return func(options *BoundedPoolOptions[D]) {
		options.StopTimeout = stopTimeout
	}
}

// WithBoundedPoolLogger настраивает worker pool logger
func WithBoundedPoolLogger[D any](logger logger.Logger) BoundedPoolOption[D] {
	return func(options *BoundedPoolOptions[D]) {
		options.Logger = logger
	}
}

// WithBoundedPoolJobHandler настраивает обработчик
func WithBoundedPoolJobHandler[D any](handler PoolJobHandler[D]) BoundedPoolOption[D] {
	return func(options *BoundedPoolOptions[D]) {
		options.JobHandler = handler
	}
}
