package tailcutter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
)

const (
	workerNameTemplate string = "tail-cutter-%s"
)

// Worker реализует контракты планировщика и воркера, оркестрируя асинхронный конвейер очистки исторического мусора.
//
// Инкапсулирует Use Case слои поиска и удаления, анонимно внедряет BaseSchedulerDispatcher ядра фреймворка,
// обеспечивает строгую защиту методов жизненного цикла от сокрытия при эмбеддинге и распараллеливает задачи.
type Worker struct {
	*worker.BaseSchedulerDispatcher[string]
	name      string
	opts      *Options
	tailGetUC usecase.TailGetUseCase[string]
	tailCutUC usecase.TailCutUseCase[string]
	logger    logger.Logger
}

// Контролируем реализацию контрактов жизненного цикла и запуска фреймворка на этапе компиляции
var _ worker.Scheduler = (*Worker)(nil)
var _ worker.CommonWorker = (*Worker)(nil)
var _ container.Runner = (*Worker)(nil)

// New — фабричный конструктор планировщика очистки хвостов Outbox-журнала.
// Выполняет сборку Functional Options, проверку инвариантов и инжектирует методы dataProvider/cutTail в базовый диспетчер.
func New(options ...Option) (*Worker, error) {
	opts := NewTailCutterOptions()
	for _, option := range options {
		option(opts)
	}
	if err := opts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewTailCutter", "tail cutter options validation failed", err)
	}
	// instance
	res := &Worker{
		name:      fmt.Sprintf(workerNameTemplate, opts.Name),
		opts:      opts,
		tailGetUC: opts.TailGetUC,
		tailCutUC: opts.TailCutUC,
		logger:    opts.Logger.GetLogger(fmt.Sprintf(workerNameTemplate, opts.Name)),
	}
	// scheduler dispatcher
	dispatcher, err := worker.NewBaseSchedulerDispatcher[string](
		worker.WithSchedulerDispatcherName[string](opts.Name),
		worker.WithSchedulerDispatcherStopTimeout[string](opts.StopTimeout),
		worker.WithSchedulerDispatcherDataProvider[string](res.dataProvider),
		worker.WithSchedulerDispatcherLogger[string](opts.Logger),
		worker.WithSchedulerDispatcherPoolWorkerCount[string](opts.WorkerCount),
		worker.WithSchedulerDispatcherPoolDataCapacity[string](opts.DataCapacity),
		worker.WithSchedulerDispatcherPoolCompleteProcess[string](opts.CompleteProcess),
		worker.WithSchedulerDispatcherPoolJobHandler[string](res.cutTail),
		worker.WithSchedulerDispatcherSchedulerStartInterval[string](opts.StartInterval),
		worker.WithSchedulerDispatcherSchedulerScheduleInterval[string](opts.ScheduleInterval),
	)
	if err != nil {
		return nil, errs.NewTlCommonError("NewTailCutter", "tail cutter scheduler dispatcher creation failed", err)
	}
	// setup
	res.BaseSchedulerDispatcher = dispatcher

	return res, nil
}

// Start инициализирует и запускает внутренний цикл планировщика базового диспетчера.
// Явное объявление метода на уровне структуры защищает публичный SDK-компонент от случайного сокрытия (shadowing) при эмбеддинге.
func (tc *Worker) Start(ctx context.Context) error {
	return tc.BaseSchedulerDispatcher.Start(ctx)
}

// Stop выполняет мягкую остановку планировщика, сигнализирует контексту и ожидает завершения активных воркеров.
func (tc *Worker) Stop(ctx context.Context) error {
	return tc.BaseSchedulerDispatcher.Stop(ctx)
}

// IsRunning возвращает текущий статус активности внутреннего диспетчера планировщика.
func (tc *Worker) IsRunning() bool {
	return tc.BaseSchedulerDispatcher.IsRunning()
}

// dataProvider автоматически вызывается ядром шедулера по тикеру.
// Вычисляет точную временную границу отсечки и выгружает пачку ID устаревших логов для отправки в очередь пула воркеров.
func (tc *Worker) dataProvider(ctx context.Context, eventTime time.Time) ([]string, error) {
	tc.GetLogger().Debugf("tail cutter %s time event %s data provider start", tc.GetName(), eventTime.Format(time.DateTime))
	defer tc.GetLogger().Debugf("tail cutter %s time event %s data provider finish", tc.GetName(), eventTime.Format(time.DateTime))

	tailCutTime := eventTime.Add(-tc.opts.DataInterval)
	tc.GetLogger().Debugf("tail cutter %s time event %s data provider tail time %s ", tc.GetName(), eventTime.Format(time.DateTime), tailCutTime.Format(time.DateTime))

	if !tc.opts.CutEnabled {
		tc.GetLogger().Debugf("tail cutter %s time event %s tail cut enabled [%v] pass iteration", tc.GetName(), eventTime.Format(time.DateTime), tc.opts.CutEnabled)

		return []string{}, nil
	}

	res, err := tc.tailGetUC.GetTail(ctx, tailCutTime)
	if err != nil {
		return nil, err
	}

	tc.GetLogger().Debugf("tail cutter %s time event %s total data records to dispatch [%v]", tc.GetName(), eventTime.Format(time.DateTime), len(res))

	return res, nil
}

// cutTail выступает в роли обработчика задачи (Job Handler) внутри пула воркеров.
// Распараллеливает точечное транзакционное удаление конкретного ID outbox-записи.
func (tc *Worker) cutTail(ctx context.Context, workerIndex int, data string) error {
	tc.GetLogger().Debugf("tail cutter %s worker %v cut tail start", tc.GetName(), workerIndex)
	defer tc.GetLogger().Debugf("tail cutter %s worker %v cut tail finish", tc.GetName(), workerIndex)

	return tc.tailCutUC.Cut(ctx, data)
}

// GetName возвращает скомпилированное системное имя текущего шедулера.
func (tc *Worker) GetName() string {
	return tc.name
}

// GetLogger возвращает инстанс изолированного логгера шедулера.
func (tc *Worker) GetLogger() logger.Logger {
	return tc.logger
}

// GetOpts возвращает ссылку на привязанный объект конфигурационных параметров.
func (tc *Worker) GetOpts() *Options {
	return tc.opts
}

// GetContext возвращает контекст управления жизненным циклом базового диспетчера.
func (tc *Worker) GetContext() context.Context {
	return tc.BaseSchedulerDispatcher.GetContext()
}

// GetContextCancel возвращает функцию отмены контекста базового диспетчера.
func (tc *Worker) GetContextCancel() context.CancelFunc {
	return tc.BaseSchedulerDispatcher.GetContextCancel()
}

// GetWaitGroup возвращает указатель на синхронизатор WaitGroup, контролирующий завершение горутин пула воркеров.
func (tc *Worker) GetWaitGroup() *sync.WaitGroup {
	return tc.BaseSchedulerDispatcher.GetWaitGroup()
}
