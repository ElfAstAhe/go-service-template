package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// BasePool реализует интерфейсы CommonWorker, Pool и container.Runner, представляя собой
// промышленный высокопроизводительный пул конкурентных воркеров (Generic Worker Pool).
//
// Оркестрирует распределение строго типизированных задач D по пулу горутин, контролирует атомарные статусы
// жизненного цикла и предоставляет гибкие сценарии плавного тушения буферов (Graceful Shutdown).
type BasePool[D any] struct {
	name       string              // Уникальное имя пула для детализации контекстов логирования и метрик
	ctx        context.Context     // Контекст времени жизни горутин пула
	cancel     context.CancelFunc  // Функция экстренной отмены рантайма пула
	wg         sync.WaitGroup      // WaitGroup контроля физического завершения циклов всех воркеров пула
	dataChan   chan D              // Потокобезопасный буферизованный сетевой канал распределения задач
	jobHandler JobHandler[D]       // Потребительский прикладной обработчик бизнес-логики задачи
	opts       *BasePoolOptions[D] // указатель на параметры пула
	log        logger.Logger       // Изолированный структурированный логгер компонента
	running    *atomic.Bool        // Атомарный флаг активности, защищающий от двойного запуска/останова
}

// Проверяем строгое соответствие контрактам и интерфейсам на этапе компиляции
var _ CommonWorker = (*BasePool[string])(nil)
var _ Pool[string] = (*BasePool[string])(nil)
var _ container.Runner = (*BasePool[string])(nil)

// NewBasePool - конструктор дженерик-пула воркеров.
func NewBasePool[D any](options ...BasePoolOption[D]) (*BasePool[D], error) {
	opts := NewBasePoolOptions[D]()

	for _, opt := range options {
		opt(opts)
	}

	if err := opts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewBasePool[D]", "validate options failed", err)
	}

	res := &BasePool[D]{
		name:       opts.Name,
		jobHandler: opts.JobHandler,
		opts:       opts,
		log:        opts.Logger.GetLogger(opts.Name),
		running:    new(atomic.Bool),
	}
	res.running.Store(false)

	return res, nil
}

// Start осуществляет атомарный запуск пула горутин-обработчиков (FIFO распределение).
func (bp *BasePool[D]) Start(ctx context.Context) error {
	// Использование CAS операции исключает гонки данных при повторных или параллельных вызовах Start
	if !bp.running.CompareAndSwap(false, true) {
		return errs.NewCommonError(fmt.Sprintf("worker pool %s already started", bp.GetName()), nil)
	}

	bp.GetLogger().Debugf("worker pool %s starting", bp.GetName())
	defer bp.GetLogger().Debugf("worker pool %s started", bp.GetName())

	// Инициализируем контекст пула на базе родительского контекста приложения
	//nolint:gosec // G118 :
	bp.ctx, bp.cancel = context.WithCancel(ctx)

	// Guard Clause against context leaks (Gosec G118 fix)
	isStartedSuccessfully := false
	defer func() {
		if !isStartedSuccessfully && bp.cancel != nil {
			bp.cancel()
		}
	}()

	// Аллоцируем буферизованный канал задач согласно лимитам DataCapacity
	bp.dataChan = make(chan D, bp.GetOpts().DataCapacity)

	// Конкурентно разворачиваем веер фиксированного количества горутин-воркеров
	for i := 0; i < bp.GetOpts().WorkerCount; i++ {
		bp.GetWaitGroup().Add(1)
		go bp.worker(i)
	}

	return nil
}

// Stop выполняет фазу контролируемой мягкой остановки пула с верификацией флагов завершения буфера.
func (bp *BasePool[D]) Stop(stopCtx context.Context) error {
	if !bp.running.CompareAndSwap(true, false) {
		return errs.NewCommonError(fmt.Sprintf("worker pool %s not running", bp.GetName()), nil)
	}

	bp.GetLogger().Debugf("worker pool %s stopping", bp.GetName())
	defer bp.GetLogger().Debugf("worker pool %s stopped", bp.GetName())

	// 1. Закрываем канал: новые Push/TryPush начнут отсекаться, воркеры увидят закрытие (!opened)
	close(bp.dataChan)

	// 2. Стратегия Fast-Shutdown: если завершение буфера не требуется — принудительно гасим контекст
	if !bp.GetOpts().CompleteProcess && bp.GetContextCancel() != nil {
		bp.GetLogger().Debugf("worker pool %s is not complete data channel processing, cancel pool context", bp.GetName())
		bp.GetContextCancel()()
	}

	// 3. Запускаем фоновый барьер ожидания гашения воркеров
	stopChan := make(chan struct{})
	go func() {
		bp.GetWaitGroup().Wait()
		close(stopChan)
	}()

	bp.GetLogger().Debugf("worker pool %s waiting for workers to stop", bp.GetName())
	// 4. Селекторный барьер контроля жестких лимитов таймаута останова
	select {
	case <-stopChan:
		bp.GetLogger().Debugf("worker pool %s stopped gracefully, all data processed", bp.GetName())
	case <-time.After(bp.opts.StopTimeout):
		// Защита от вечного зависания: выходим по локальному лимиту времени пула
		bp.GetLogger().Debugf("worker pool %s stop timed out, force stopping, some data not processed and will be lost", bp.GetName())
	case <-stopCtx.Done():
		// Защита от вечного зависания: выходим по общему верхнему лимиту контекста приложения
		bp.GetLogger().Debugf("worker pool %s stopped by stop context, force stopping, some data not processed and will be lost", bp.GetName())
	}

	// Выполняем финальный сброс контекста
	if bp.GetContextCancel() != nil {
		bp.GetContextCancel()()
	}

	return nil
}

// Push выполняет блокирующую вставку задачи в очередь пула с защитой контекста выполнения.
func (bp *BasePool[D]) Push(data D) {
	if !bp.IsRunning() {
		return
	}

	select {
	case bp.dataChan <- data:
		bp.GetLogger().Debugf("worker pool %s push data [%v]", bp.GetName(), data)
	case <-bp.GetContext().Done():
		bp.GetLogger().Debugf("worker pool %s stop push by context", bp.GetName())
	}
}

// TryPush выполняет мгновенную неблокирующую попытку вставки задачи с мгновенным возвратом статуса успеха.
func (bp *BasePool[D]) TryPush(data D) bool {
	if !bp.IsRunning() {
		return false
	}

	select {
	case bp.dataChan <- data:
		bp.GetLogger().Debugf("worker pool %s push data [%v]", bp.GetName(), data)
		return true
	case <-bp.GetContext().Done():
		bp.GetLogger().Debugf("worker pool %s stop push by context", bp.GetName())
		return false
	default:
		// Сценарий перегрузки (Buffer Overflow): буфер полон, таска игнорируется во избежание блокировки I/O
		bp.GetLogger().Debugf("worker pool %s try push default, data [%v] ignored and lost", bp.GetName(), data)
		return false
	}
}

// Len возвращает текущую фактическую длину заполненности внутреннего канала задач.
func (bp *BasePool[D]) Len() int {
	return len(bp.dataChan)
}

// Capacity возвращает верхний жесткий лимит вместимости канала.
func (bp *BasePool[D]) Capacity() int {
	return cap(bp.dataChan)
}

// Внутренний бесконечный цикл горутины-обработчика пула воркеров (Event Loop)
func (bp *BasePool[D]) worker(workerIndex int) {
	bp.GetLogger().Debugf("worker pool %s worker %v start", bp.GetName(), workerIndex)
	defer bp.GetLogger().Debugf("worker pool %s worker %v finish", bp.GetName(), workerIndex)
	defer bp.GetWaitGroup().Done()

	for {
		select {
		case <-bp.GetContext().Done():
			bp.GetLogger().Debugf("worker pool %s worker %v context done, stop worker", bp.GetName(), workerIndex)
			return
		case data, opened := <-bp.dataChan:
			// Перехватываем закрытие канала: если канал закрыт и буфер пуст — мягко завершаем горутину
			if !opened {
				bp.GetLogger().Debugf("worker pool %s worker %v queue closed, stop worker", bp.GetName(), workerIndex)
				return
			}
			// Запуск прикладной бизнес-логики обработчика
			if bp.jobHandler != nil {
				err := bp.jobHandler(bp.GetContext(), workerIndex, data)
				if err != nil {
					bp.GetLogger().Errorf("worker pool %s worker %v job failed:  %v", bp.GetName(), workerIndex, err)
				}
			} else {
				bp.GetLogger().Warnf("worker pool %s worker %v job handler not applied, data lost", bp.GetName(), workerIndex)
			}
		}
	}
}

// GetName возвращает уникальное текстовое наименование текущего экземпляра пула воркеров.
func (bp *BasePool[D]) GetName() string {
	return bp.name
}

// GetContext возвращает ссылку на внутренний контекст context.Context времени жизни пула.
func (bp *BasePool[D]) GetContext() context.Context {
	return bp.ctx
}

// GetContextCancel возвращает функцию-триггер context.CancelFunc для принудительной отмены контекста пула.
func (bp *BasePool[D]) GetContextCancel() context.CancelFunc {
	return bp.cancel
}

// GetLogger возвращает инстанс изолированного структурированного логгера, закрепленный за пулом.
func (bp *BasePool[D]) GetLogger() logger.Logger {
	return bp.log
}

// GetWaitGroup возвращает указатель на общую структуру sync.WaitGroup контроля запущенных дочерних горутин.
func (bp *BasePool[D]) GetWaitGroup() *sync.WaitGroup {
	return &bp.wg
}

// GetOpts возвращает ссылку на опции пула BasePoolOptions.
func (bp *BasePool[D]) GetOpts() *BasePoolOptions[D] {
	return bp.opts
}

// IsRunning возвращает текущий атомарный статус активности пула (true — запущен и принимает задачи).
func (bp *BasePool[D]) IsRunning() bool {
	return bp.running.Load()
}
