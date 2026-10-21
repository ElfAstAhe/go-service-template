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
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

const boundedPoolNameTemplate string = "bounded-pool-%s"

// BoundedPool реализует интерфейсы Common, Pool и container.Runner, представляя собой
// промышленный высокопроизводительный пул конкурентных воркеров (Generic Worker Pool).
//
// Оркестрирует распределение строго типизированных задач D по пулу горутин, контролирует атомарные статусы
// жизненного цикла и предоставляет гибкие сценарии плавного тушения буферов (Graceful Shutdown).
type BoundedPool[D any] struct {
	name       string                 // Уникальное имя пула для детализации контекстов логирования и метрик
	ctx        context.Context        // Контекст времени жизни горутин пула
	cancel     context.CancelFunc     // Функция экстренной отмены рантайма пула
	wg         sync.WaitGroup         // WaitGroup контроля физического завершения циклов всех воркеров пула
	dataChan   chan D                 // Потокобезопасный буферизованный сетевой канал распределения задач
	jobHandler PoolJobHandler[D]      // Потребительский прикладной обработчик бизнес-логики задачи
	opts       *BoundedPoolOptions[D] // указатель на параметры пула
	log        logger.Logger          // Изолированный структурированный логгер компонента
	running    *atomic.Bool           // Атомарный флаг активности, защищающий от двойного запуска/останова
}

// Проверяем строгое соответствие контрактам и интерфейсам на этапе компиляции
var _ Common = (*BoundedPool[string])(nil)
var _ Pool[string] = (*BoundedPool[string])(nil)
var _ container.Runner = (*BoundedPool[string])(nil)

// NewBoundedPool - конструктор дженерик-пула воркеров.
func NewBoundedPool[D any](options ...BoundedPoolOption[D]) (*BoundedPool[D], error) {
	opts := NewBoundedPoolOptions[D]()
	for _, opt := range options {
		opt(opts)
	}
	if err := opts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewBoundedPool[D]", "validate options failed", err)
	}

	res := &BoundedPool[D]{
		name:       fmt.Sprintf(boundedPoolNameTemplate, opts.Name),
		jobHandler: opts.JobHandler,
		opts:       opts,
		log:        opts.Logger.GetLogger(fmt.Sprintf(boundedPoolNameTemplate, opts.Name)),
		running:    new(atomic.Bool),
	}
	res.running.Store(false)

	return res, nil
}

// Start осуществляет атомарный запуск пула горутин-обработчиков (FIFO распределение).
func (bp *BoundedPool[D]) Start(ctx context.Context) error {
	// Использование CAS операции исключает гонки данных при повторных или параллельных вызовах Start
	if !bp.running.CompareAndSwap(false, true) {
		//		return errs.NewCommonError(fmt.Sprintf("worker pool %s already started", bp.GetName()), nil)

		return nil
	}

	bp.GetLogger().Debugf("bounded pool %s starting", bp.GetName())
	defer bp.GetLogger().Debugf("bounded pool %s started", bp.GetName())

	// Инициализируем контекст пула на базе родительского контекста приложения
	//nolint:gosec // G118 : worker background context
	bp.ctx, bp.cancel = context.WithCancel(ctx)

	// Аллоцируем буферизованный канал задач согласно лимитам DataCapacity
	bp.dataChan = make(chan D, bp.GetOpts().DataCapacity)

	// Конкурентно разворачиваем веер фиксированного количества горутин-воркеров
	for i := 0; i < bp.GetOpts().WorkerCount; i++ {
		bp.wg.Add(1)
		go bp.worker(i)
	}

	return nil
}

// Stop выполняет фазу контролируемой мягкой остановки пула с верификацией флагов завершения буфера.
func (bp *BoundedPool[D]) Stop(stopCtx context.Context) error {
	if !bp.running.CompareAndSwap(true, false) {
		//		return errs.NewCommonError(fmt.Sprintf("worker pool %s not running", bp.GetName()), nil)

		return nil
	}

	bp.GetLogger().Debugf("bounded pool %s stopping", bp.GetName())
	defer bp.GetLogger().Debugf("bounded pool %s stopped", bp.GetName())

	// 1. Закрываем канал: новые Push/TryPush начнут отсекаться, воркеры увидят закрытие (!opened)
	close(bp.dataChan)

	// 2. Стратегия Fast-Shutdown: если завершение буфера не требуется — принудительно гасим контекст
	if !bp.GetOpts().CompleteProcess && bp.cancel != nil {
		bp.GetLogger().Debugf("bounded pool %s is not complete data channel processing, cancel pool context", bp.GetName())
		bp.cancel()
	}

	// 3. Запускаем фоновый барьер ожидания гашения воркеров
	stopChan := make(chan struct{})
	go func() {
		bp.wg.Wait()
		close(stopChan)
	}()

	bp.GetLogger().Debugf("bounded pool %s waiting for workers to stop", bp.GetName())
	// 4. Селекторный барьер контроля жестких лимитов таймаута останова
	select {
	case <-stopChan:
		bp.GetLogger().Debugf("bounded pool %s stopped gracefully, all data processed", bp.GetName())
	case <-time.After(bp.opts.StopTimeout):
		// Защита от вечного зависания: выходим по локальному лимиту времени пула
		bp.GetLogger().Debugf("bounded pool %s stop timed out, force stopping, some data not processed and will be lost", bp.GetName())
	case <-stopCtx.Done():
		// Защита от вечного зависания: выходим по общему верхнему лимиту контекста приложения
		bp.GetLogger().Debugf("bounded pool %s stopped by stop context, force stopping, some data not processed and will be lost", bp.GetName())
	}

	// Выполняем финальный сброс контекста
	if bp.cancel != nil {
		bp.cancel()
	}

	return nil
}

// GetName возвращает уникальное текстовое наименование текущего экземпляра пула воркеров.
func (bp *BoundedPool[D]) GetName() string {
	return bp.name
}

// IsRunning возвращает текущий атомарный статус активности пула (true — запущен и принимает задачи).
func (bp *BoundedPool[D]) IsRunning() bool {
	return bp.running.Load()
}

// Push выполняет атомарную блокирующую вставку задачи в очередь пула с защитой контекста выполнения.
// Возвращает ошибку, если пул остановлен, отменен внешний контекст или завершился жизненный цикл воркера.
func (bp *BoundedPool[D]) Push(ctx context.Context, data D) error {
	bp.GetLogger().Debugf("bounded pool %s Push start", bp.GetName())
	defer bp.GetLogger().Debugf("bounded pool %s Push finish", bp.GetName())

	if !bp.IsRunning() {
		return errs.NewTlCommonError("BoundedPool[D].Push", "worker not running", nil)
	}

	select {
	case bp.dataChan <- data:
		bp.GetLogger().Debugf("bounded pool %s push data [%v]", bp.GetName(), data)
		return nil
	case <-ctx.Done():
		bp.GetLogger().Debugf("bounded pool %s stop push by context", bp.GetName())
		return errs.NewTlCommonError("BoundedPool[D].Push", "stop push by context", ctx.Err())
	case <-bp.GetContext().Done():
		bp.GetLogger().Debugf("bounded pool %s stop push by worker context", bp.GetName())
		return errs.NewTlCommonError("BoundedPool[D].Push", "stop push by worker context", ctx.Err())
	}
}

// TryPush выполняет мгновенную неблокирующую попытку вставки задачи с мгновенным возвратом статуса успеха.
// Возвращает true при успешном добавлении, и false при перегрузке буфера или отмене контекстов.
func (bp *BoundedPool[D]) TryPush(ctx context.Context, data D) bool {
	bp.GetLogger().Debugf("bounded pool %s TryPush start", bp.GetName())
	defer bp.GetLogger().Debugf("bounded pool %s TryPush finish", bp.GetName())

	if !bp.IsRunning() {
		return false
	}

	select {
	case bp.dataChan <- data:
		bp.GetLogger().Debugf("bounded pool %s push data [%v]", bp.GetName(), data)
		return true
	case <-ctx.Done():
		bp.GetLogger().Debugf("bounded pool %s stop push by context", bp.GetName())
		return false
	case <-bp.GetContext().Done():
		bp.GetLogger().Debugf("bounded pool %s stop push by worker context", bp.GetName())
		return false
	default:
		// Сценарий перегрузки (Buffer Overflow): буфер полон, таска игнорируется во избежание блокировки I/O
		bp.GetLogger().Debugf("bounded pool %s try push default, data [%v] ignored and lost", bp.GetName(), data)
		return false
	}
}

// Len возвращает текущую фактическую длину заполненности внутреннего канала задач.
func (bp *BoundedPool[D]) Len() int {
	return len(bp.dataChan)
}

// Capacity возвращает верхний жесткий лимит вместимости канала.
func (bp *BoundedPool[D]) Capacity() int {
	return cap(bp.dataChan)
}

// Внутренний бесконечный цикл горутины-обработчика пула воркеров (Event Loop)
func (bp *BoundedPool[D]) worker(workerIndex int) {
	bp.GetLogger().Debugf("bounded pool %s worker %v start", bp.GetName(), workerIndex)
	defer bp.GetLogger().Debugf("bounded pool %s worker %v finish", bp.GetName(), workerIndex)
	defer bp.wg.Done()

	for {
		select {
		case <-bp.GetContext().Done():
			bp.GetLogger().Debugf("bounded pool %s worker %v context done, stop worker", bp.GetName(), workerIndex)
			return
		case data, opened := <-bp.dataChan:
			// Перехватываем закрытие канала: если канал закрыт и буфер пуст — мягко завершаем горутину
			if !opened {
				bp.GetLogger().Debugf("bounded pool %s worker %v queue closed, stop worker", bp.GetName(), workerIndex)
				return
			}
			// Запуск прикладной бизнес-логики обработчика
			if !utils.IsNil(bp.jobHandler) {
				err := bp.jobHandler(bp.GetContext(), workerIndex, data)
				if err != nil {
					bp.GetLogger().Errorf("bounded pool %s worker %v job failed:  %v", bp.GetName(), workerIndex, err)
				}
			} else {
				bp.GetLogger().Warnf("bounded pool %s worker %v job handler not applied, data lost", bp.GetName(), workerIndex)
			}
		}
	}
}

// GetContext возвращает ссылку на внутренний контекст context.Context времени жизни пула.
func (bp *BoundedPool[D]) GetContext() context.Context {
	return bp.ctx
}

// GetLogger возвращает инстанс изолированного структурированного логгера, закрепленный за пулом.
func (bp *BoundedPool[D]) GetLogger() logger.Logger {
	return bp.log
}

// GetOpts возвращает ссылку на опции пула BoundedPoolOptions.
func (bp *BoundedPool[D]) GetOpts() *BoundedPoolOptions[D] {
	return bp.opts
}
