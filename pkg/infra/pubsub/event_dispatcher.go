package pubsub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// DefaultNotifyTimeout определяет жесткий временной потолок на обработку события одним наблюдателем.
const (
	DefaultNotifyTimeout = 5 * time.Second
)

// EventDispatcher реализует интерфейс Publisher[T], обеспечивая конкурентную, асинхронную
// и отказоустойчивую веерную рассылку (Fan-Out) бизнес-событий зарегистрированным наблюдателям.
type EventDispatcher[T any] struct {
	mu            sync.RWMutex           // RWMutex защищает потокобезопасный реестр мапы подписчиков
	name          string                 // Уникальное строковое имя диспетчера (для Sub-Logger изоляции)
	observers     map[string]Observer[T] // Индексированный реестр активных наблюдателей
	notifyTimeout time.Duration          // Индивидуальный лимит времени на обработку уведомления
	logger        logger.Logger          // Изолированный контекстный логгер диспетчера
}

// Проверяем соответствие интерфейсу Publisher для базового строкового типа на этапе компиляции
var _ Publisher[string] = (*EventDispatcher[string])(nil)

// NewEventDispatcher — фабричный конструктор диспетчера событий.
// Автоматически нарезает дочерний логгер и подкладывает дефолтный таймаут в 5 секунд при невалидных параметрах.
func NewEventDispatcher[T any](name string, notifyTimeout time.Duration, log logger.Logger) *EventDispatcher[T] {
	res := &EventDispatcher[T]{
		name:          name,
		notifyTimeout: notifyTimeout,
		logger:        log.GetLogger(name),
		observers:     make(map[string]Observer[T]),
	}
	if res.notifyTimeout <= 0 {
		res.notifyTimeout = DefaultNotifyTimeout
	}

	return res
}

// Register атомарно добавляет наблюдателя в реестр рассылки под эксклюзивным локом (Write Lock).
func (ed *EventDispatcher[T]) Register(observer Observer[T]) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	if utils.IsNil(observer) {
		return
	}

	ed.observers[observer.GetName()] = observer
}

// Unregister атомарно исключает наблюдателя из реестра рассылки по его уникальному имени.
func (ed *EventDispatcher[T]) Unregister(observer Observer[T]) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	if utils.IsNil(observer) {
		return
	}

	delete(ed.observers, observer.GetName())
}

// Notify выполняет мгновенный снимок (Snapshot) пула подписчиков под RLock и порождает
// фоновую горутину для полностью асинхронного, неблокирующего основной поток выполнения рассылки.
func (ed *EventDispatcher[T]) Notify(ctx context.Context, data T) {
	ed.logger.Debugf("pub/sub event dispatcher %s Notify start", ed.GetName())
	defer ed.logger.Debugf("pub/sub event dispatcher %s Notify finish", ed.GetName())

	ed.mu.RLock()
	if len(ed.observers) == 0 {
		ed.mu.RUnlock()
		return
	}

	// Делаем мгновенный слепок подписчиков, минимизируя время удержания RLock мьютекса
	observers := make([]Observer[T], 0, len(ed.observers))
	for _, observer := range ed.observers {
		observers = append(observers, observer)
	}
	ed.mu.RUnlock()

	// 💡 Архитектурный паттерн (Fire-and-Forget): context.WithoutCancel(ctx) отвязывает фоновую горутину
	// от отмены родительского контекста (например, завершения HTTP-запроса), сохраняя при этом все TraceID!
	go ed.internalNotify(context.WithoutCancel(ctx), data, observers)
}

// internalNotify запускает параллельную веерную обработку (Fan-Out Pattern) события.
// Каждый обсервер изолируется в своей горутине, защищенной персональным тайм-лимитом и перехватом паник.
func (ed *EventDispatcher[T]) internalNotify(ctx context.Context, data T, observers []Observer[T]) {
	ed.logger.Debugf("pub/sub event dispatcher %s internalNotify start", ed.GetName())
	defer ed.logger.Debugf("pub/sub event dispatcher %s internalNotify finish", ed.GetName())

	var wg sync.WaitGroup
	asyncCtx, asyncCancel := context.WithTimeout(ctx, ed.notifyTimeout)
	defer asyncCancel()

	for _, observer := range observers {
		obs := observer
		wg.Add(1)

		// Порождаем конкурентный воркер под каждого зарегистрированного обсервера
		go func(observe Observer[T]) {
			ed.logger.Debugf("pub/sub event dispatcher %s observer %s start", ed.GetName(), observe.GetName())
			defer ed.logger.Debugf("pub/sub event dispatcher %s observer %s finish", ed.GetName(), observe.GetName())
			defer wg.Done()

			// Защитный барьер: локальный recover() спасает диспетчер от падения, если обсервер выкинет рантайм-панику
			defer func() {
				if r := recover(); r != nil {
					var recoveryErr error
					if e, ok := r.(error); ok {
						recoveryErr = errs.NewCommonError("panic recovery", e)
					} else {
						recoveryErr = errs.NewCommonError(fmt.Sprintf("panic recovery [%v]", r), nil)
					}
					ed.logger.Errorf("pub/sub event dispatcher %s observer %s panic recovery %v", ed.GetName(), observe.GetName(), recoveryErr)
				}
			}()

			// Выполняем целевое уведомление подписчика с передачей асинхронного защищенного контекста
			if err := observe.OnNotify(asyncCtx, data); err != nil {
				ed.logger.Errorf("pub/sub event dispatcher %s observer %s on notify got error %v", ed.GetName(), observe.GetName(), err)
			}
		}(obs)
	}

	// Ожидаем завершения или таймаута всей пачки горутин-наблюдателей
	wg.Wait()
}

// GetName возвращает строковый идентификатор текущего диспетчера.
func (ed *EventDispatcher[T]) GetName() string {
	return ed.name
}
