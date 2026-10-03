package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const schedulerDispatcherNameTemplate string = "scheduler-dispatcher-%s"

// BaseSchedulerDispatcher реализует интерфейсы CommonWorker, Scheduler и container.Runner,
// являясь строго типизированным оркестратором периодических задач (Generic Scheduler Dispatcher Pattern).
//
// Инкапсулирует под своим крылом базовый таймер (BaseScheduler) и внутренний многопоточный пул воркеров (Pool).
// За счет асинхронного веерного распределения задач полностью защищает системные таймеры Go от дрейфа времени,
// изолируя фазу извлечения данных (Data Fetching) от тяжелой параллельной бизнес-обработки (Job Processing).
type BaseSchedulerDispatcher[D comparable] struct {
	ctx          context.Context
	cancel       context.CancelFunc
	scheduler    *BaseScheduler                     // Встраивание (Embedding) базового планировщика для управления событиями таймера
	workerPool   Pool[D]                            // Ссылка на строго типизированный внутренний пул конкурентных воркеров
	opts         *BaseSchedulerDispatcherOptions[D] // Указатель на агрегированную структуру конфигурационных параметров
	dataProvider DispatcherDataProvider[D]          // Провайдер-поставщик пачек данных для обработки по расписанию
}

// Гарантируем строгое соответствие контрактам и интерфейсам фреймворка на этапе компиляции
var _ CommonWorker = (*BaseSchedulerDispatcher[string])(nil)
var _ Scheduler = (*BaseSchedulerDispatcher[string])(nil)
var _ container.Runner = (*BaseSchedulerDispatcher[string])(nil)

// NewBaseSchedulerDispatcher — фабричный конструктор строго типизированного комплексного диспетчера.
// Автоматически инициализирует внутренний BasePool и связывает его с циклом событий встроенного планировщика.
func NewBaseSchedulerDispatcher[D comparable](options ...BaseSchedulerDispatcherOption[D]) (*BaseSchedulerDispatcher[D], error) {
	opts := NewBaseSchedulerDispatcherOptions()
	for _, opt := range options {
		opt(opts)
	}
	if err := opts.Validate(); err != nil {
		panic(err)
	}

	res := &BaseSchedulerDispatcher[D]{
		name:         fmt.Sprintf(schedulerDispatcherNameTemplate, opts.Name),
		dataProvider: dataProvider,
		workerPool:   NewBasePool[D](name, config.PoolConfig, jobHandler, log),
	}

	// Инициализируем базовый планировщик, подставляя в качестве колбека внутренний метод timerDispatcher
	res.BaseScheduler = NewBaseScheduler(name, res.timerDispatcher, config.SchedulerConfig, log)

	return res
}

// Start выполняет каскадный запуск рантайма: сначала переводит в активное состояние внутренний пул воркеров,
// а затем взводит циклический таймер планировщика.
// В случае сбоя старта любого из компонентов инициирует экстренный автоматический откат (Stop)
// в рамках жесткого лимита времени (100мс) и возвращает агрегированную системную ошибку CommonError.
func (bsd *BaseSchedulerDispatcher[D]) Start(ctx context.Context) error {
	errPool := bsd.workerPool.Start(ctx)
	errScheduler := bsd.BaseScheduler.Start(ctx)
	err := errors.Join(errPool, errScheduler)
	if err != nil {
		// Фаза автоматического Fail-Fast отката: принудительно гасим частично поднятые ресурсы
		stopCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err = errors.Join(err, bsd.Stop(stopCtx))

		return errs.NewCommonError(fmt.Sprintf("scheduler dispatcher %s start failed", bsd.GetName()), err)
	}

	return nil
}

// Stop осуществляет синхронную мягкую остановку (Graceful Shutdown) всей экосистемы диспетчера.
// Сначала деактивирует и очищает системный таймер в ядре ОС, блокируя генерацию новых событий,
// а затем каскадно останавливает внутренний пул воркеров, давая им доварить текущие задачи.
// Агрегирует ошибки останова обоих узлов через errors.Join, возвращая CommonError при сбоях.
func (bsd *BaseSchedulerDispatcher[D]) Stop(stopCtx context.Context) error {
	errScheduler := bsd.BaseScheduler.Stop(stopCtx)
	errPool := bsd.workerPool.Stop(stopCtx)
	err := errors.Join(errPool, errScheduler)
	if err != nil {
		return errs.NewCommonError(fmt.Sprintf("scheduler dispatcher %s stop failed", bsd.GetName()), err)
	}

	return nil
}

// timerDispatcher — внутренний (неэкспортируемый) мост-колбек, связывающий тики таймера с очередью пула.
// Вызывается базовым планировщиком, извлекает пачку данных через dataProvider и веерно распределяет их по воркерам.
func (bsd *BaseSchedulerDispatcher[D]) timerDispatcher(ctx context.Context, eventTime time.Time) error {
	bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s start", bsd.GetName(), eventTime.Format(time.DateTime))
	defer bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s finish", bsd.GetName(), eventTime.Format(time.DateTime))

	if bsd.dataProvider == nil {
		return errs.NewCommonError(fmt.Sprintf("scheduler dispatcher %s time event %s data provider not applied", bsd.GetName(), eventTime.Format(time.DateTime)), nil)
	}

	// Извлекаем массив задач из внешнего источника (БД, кэш, API) средствами прикладного провайдера
	res, err := bsd.dataProvider(bsd.GetContext(), eventTime)
	if err != nil {
		return err
	}
	bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s got %v data records", bsd.GetName(), eventTime.Format(time.DateTime), len(res))

	// Итерируемся по пачке извлеченных данных для их маршутизации в очередь воркеров
	for _, data := range res {
		select {
		case <-ctx.Done():
			// 🛡️ Защитный барьер: если контекст отменен (SIGTERM), немедленно прекращаем раздачу задач пулу
			bsd.GetLogger().Debugf("scheduler dispatcher %s time event %s break by context done", bsd.GetName(), eventTime.Format(time.DateTime))

			return ctx.Err()
		default:
			// Блокирующий или неблокирующий (зависит от настроек канала) сброс задачи в буфер пула воркеров
			bsd.workerPool.Push(data)
		}
	}

	return nil
}

// GetOpts возвращает ссылку на конфигурационный паспорт агрегированных параметров диспетчера BaseSchedulerDispatcherConfig.
func (bsd *BaseSchedulerDispatcher[D]) GetOpts() *BaseSchedulerDispatcherOptions[D] {
	return bsd.opts
}
