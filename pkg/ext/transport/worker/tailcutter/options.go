package tailcutter

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// Константы со значениями по умолчанию для конфигурации планировщика Worker.
const (
	defaultTailCutterWorkerCount      int           = 1
	defaultTailCutterDataCapacity     int           = 32
	defaultTailCutterCompleteProcess  bool          = false
	defaultTailCutterStartInterval    time.Duration = 30 * time.Second
	defaultTailCutterScheduleInterval time.Duration = 5 * time.Minute
	defaultTailCutterDataInterval     time.Duration = 14 * 24 * time.Hour
	defaultTailCutterCutEnabled       bool          = false
)

// Option определяет сигнатуру функции для настройки параметров Worker.
type Option func(*Options)

// Options содержит конфигурационные параметры для планировщика обрезки хвостов данных.
// Включает в себя базовые опции диспетчера через эмбеддинг BaseSchedulerDispatcherOptions.
type Options struct {
	*worker.BaseSchedulerDispatcherOptions[string]
	TailGetUC    usecase.TailGetUseCase[string]
	TailCutUC    usecase.TailCutUseCase[string]
	DataInterval time.Duration
	CutEnabled   bool
}

// NewTailCutterOptions создает и возвращает новый экземпляр Options
// с базовыми настройками диспетчера и значениями интервалов по умолчанию.
func NewTailCutterOptions() *Options {
	res := &Options{
		BaseSchedulerDispatcherOptions: tailCutterBuildDefaultDispatcherOptions(),
		DataInterval:                   defaultTailCutterDataInterval,
		CutEnabled:                     defaultTailCutterCutEnabled,
	}

	return res
}

// Validate проверяет корректность заполнения всех обязательных полей конфигурации,
// включая встроенные параметры базового диспетчера и специфичные параметры use case.
//
//goland:noinspection DuplicatedCode
func (tco *Options) Validate() error {
	if strings.TrimSpace(tco.Name) == "" {
		return errs.NewTlCommonError("Validate", "name is required", nil)
	}
	if tco.WorkerCount <= 0 {
		return errs.NewTlCommonError("Validate", "worker count is required", nil)
	}
	if tco.DataCapacity <= 0 {
		return errs.NewTlCommonError("Validate", "data capacity is required", nil)
	}
	if tco.StopTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "stop timeout is required", nil)
	}
	//if utils.IsNil(tco.JobHandler) {
	//	return errs.NewTlCommonError("Validate", "job handler is required", nil)
	//}
	if utils.IsNil(tco.Logger) {
		return errs.NewTlCommonError("Validate", "logger is required", nil)
	}
	//if utils.IsNil(tco.DataProvider) {
	//	return errs.NewTlCommonError("Validate", "data provider is required", nil)
	//}
	if tco.StartInterval <= 0 {
		return errs.NewTlCommonError("Validate", "start interval is required", nil)
	}
	if tco.ScheduleInterval <= 0 {
		return errs.NewTlCommonError("Validate", "schedule interval is required", nil)
	}
	if utils.IsNil(tco.TailGetUC) {
		return errs.NewTlCommonError("Validate", "tail get use case is required", nil)
	}
	if utils.IsNil(tco.TailCutUC) {
		return errs.NewTlCommonError("Validate", "tail cut use case is required", nil)
	}
	if tco.DataInterval <= 0 {
		return errs.NewTlCommonError("Validate", "data interval is required", nil)
	}

	return nil
}

// WithName задает имя планировщика Worker.
func WithName(name string) Option {
	return func(options *Options) {
		options.Name = name
	}
}

// WithWorkerCount задает количество параллельных воркеров для пула обработки.
func WithWorkerCount(workerCount int) Option {
	return func(options *Options) {
		options.WorkerCount = workerCount
	}
}

// WithDataCapacity устанавливает емкость очереди данных для обработки.
func WithDataCapacity(capacity int) Option {
	return func(options *Options) {
		options.DataCapacity = capacity
	}
}

// WithCompleteProcess управляет флагом завершения всех задач в очереди при остановке.
func WithCompleteProcess(flag bool) Option {
	return func(options *Options) {
		options.CompleteProcess = flag
	}
}

// WithStartInterval определяет интервал задержки перед первым запуском планировщика.
func WithStartInterval(interval time.Duration) Option {
	return func(options *Options) {
		options.StartInterval = interval
	}
}

// WithScheduleInterval определяет интервал периодического запуска задач планировщиком.
func WithScheduleInterval(interval time.Duration) Option {
	return func(options *Options) {
		options.ScheduleInterval = interval
	}
}

// WithStopTimeout задает максимальное время ожидания завершения работы при остановке.
func WithStopTimeout(timeout time.Duration) Option {
	return func(options *Options) {
		options.StopTimeout = timeout
	}
}

// WithDataInterval устанавливает временной интервал (глубину), за который отсекаются данные.
func WithDataInterval(interval time.Duration) Option {
	return func(options *Options) {
		options.DataInterval = interval
	}
}

// WithCutEnabled включает или выключает физическое удаление/обрезание данных.
func WithCutEnabled(enabled bool) Option {
	return func(options *Options) {
		options.CutEnabled = enabled
	}
}

// WithTailGetUseCase конфигурирует Use Case для получения "хвостов" данных.
func WithTailGetUseCase(uc usecase.TailGetUseCase[string]) Option {
	return func(options *Options) {
		options.TailGetUC = uc
	}
}

// WithTailCutUseCase конфигурирует Use Case для обрезания/удаления данных.
func WithTailCutUseCase(uc usecase.TailCutUseCase[string]) Option {
	return func(options *Options) {
		options.TailCutUC = uc
	}
}

// WithLogger настраивает logger
func WithLogger(log logger.Logger) Option {
	return func(options *Options) {
		options.Logger = log
	}
}

func tailCutterBuildDefaultDispatcherOptions() *worker.BaseSchedulerDispatcherOptions[string] {
	res := worker.NewBaseSchedulerDispatcherOptions[string]()
	for _, option := range tailCutterDefaultDispatcherOptionList() {
		option(res)
	}

	return res
}

func tailCutterDefaultDispatcherOptionList() []worker.BaseSchedulerDispatcherOption[string] {
	return []worker.BaseSchedulerDispatcherOption[string]{
		worker.WithSchedulerDispatcherPoolWorkerCount[string](defaultTailCutterWorkerCount),
		worker.WithSchedulerDispatcherPoolDataCapacity[string](defaultTailCutterDataCapacity),
		worker.WithSchedulerDispatcherPoolCompleteProcess[string](defaultTailCutterCompleteProcess),
		worker.WithSchedulerDispatcherSchedulerStartInterval[string](defaultTailCutterStartInterval),
		worker.WithSchedulerDispatcherSchedulerScheduleInterval[string](defaultTailCutterScheduleInterval),
	}
}
