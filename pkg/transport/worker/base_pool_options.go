package worker

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// Константы дефолтов для внутренней защиты рантайм-компонента (независимые от пакета config)
const (
	defaultWorkerCount     int           = 2
	defaultDataCapacity    int           = 64
	defaultCompleteProcess bool          = true
	defaultStopTimeout     time.Duration = time.Second * 5
)

// BasePoolOption определяет функциональный тип для конфигурации опций (Fluent API).
type BasePoolOption[D any] func(*BasePoolOptions[D])

// BasePoolOptions содержит параметры рантайма, необходимые для безопасной сборки и работы worker pool.
type BasePoolOptions[D any] struct {
	Name            string        // Наименование воркера
	WorkerCount     int           // Количество параллельно запущенных горутин-обработчиков
	DataCapacity    int           // Буферная емкость внутреннего канала задач (Backpressure window)
	CompleteProcess bool          // Флаг: вычитывать ли буфер до конца при закрытии канала (true) или тушить экстренно (false)
	StopTimeout     time.Duration // Временной лимит (таймаут) на мягкое завершение обработки перед принудительным выходом
	Logger          logger.Logger // Логгер
	JobHandler      JobHandler[D]
}

// NewBasePoolOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewBasePoolOptions[D any]() *BasePoolOptions[D] {
	return &BasePoolOptions[D]{
		WorkerCount:     defaultWorkerCount,
		DataCapacity:    defaultDataCapacity,
		CompleteProcess: defaultCompleteProcess,
		StopTimeout:     defaultStopTimeout,
	}
}

// Validate проверяет корректность абсолютно всех опций рантайма перед сборкой Receiver.
// Защищает приложение от паник библиотеки kafka-go и некорректного поведения консьюмера.
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

	return nil
}

// ====================================================================
// Fluent API методы для сборки опций отправителя
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

func WithPoolJobHandler[D any](handler JobHandler[D]) BasePoolOption[D] {
	return func(options *BasePoolOptions[D]) {
		options.JobHandler = handler
	}
}
