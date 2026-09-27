package worker

import (
	"context"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// JobHandler определяет строго типизированную (Generic) сигнатуру функции-обработчика конкретной задачи в пуле.
// Пробрасывает workerIndex (порядковый ID потока), позволяя логике использовать изолированные локальные ресурсы.
type JobHandler[D any] func(ctx context.Context, workerIndex int, data D) error

// DispatcherDataProvider определяет сигнатуру функции-провайдера данных для периодических планировщиков (Schedulers).
// Использует ограничение comparable для возможности быстрой дедупликации или фильтрации извлекаемых пачек задач.
type DispatcherDataProvider[D comparable] func(ctx context.Context, eventTime time.Time) ([]D, error)

// CommonWorker специфицирует фундаментальный сквозной контракт фонового асинхронного воркера (Background Worker Identity).
//
// Регламентирует каноничные методы запуска/останова горутин, а также предоставляет полный рантайм-паспорт
// компонента (логирование, контексты, синхронизация через WaitGroup), необходимый IoC-сервисам для сквозного аудита здоровья.
type CommonWorker interface {
	// Start инициирует асинхронный запуск основного цикла обработки воркера (или пула воркеров) внутри горутин.
	Start(ctx context.Context) error

	// Stop запускает процедуру контролируемой мягкой остановки (Graceful Shutdown) с фиксацией таймаутов в stopCtx.
	Stop(stopCtx context.Context) error

	// GetName возвращает уникальное текстовое наименование воркера (например, "db-janitor-scheduler").
	GetName() string

	// GetContext возвращает ссылку на внутренний контекст времени жизни текущего воркера.
	GetContext() context.Context

	// GetContextCancel возвращает функцию-триггер для принудительной экстренной отмены контекста воркера.
	GetContextCancel() context.CancelFunc

	// GetLogger возвращает инстанс изолированного структурированного логгера воркера.
	GetLogger() logger.Logger

	// GetWaitGroup возвращает указатель на общую структуру sync.WaitGroup контроля завершения всех дочерних потоков.
	GetWaitGroup() *sync.WaitGroup

	// IsRunning возвращает текущий атомарный статус активности и здоровья внутренних горутин воркера.
	IsRunning() bool
}
