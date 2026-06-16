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

const (
	DefaultNotifyTimeout = 5 * time.Second
)

type EventDispatcher[T any] struct {
	mu            sync.RWMutex
	name          string
	observers     map[string]Observer[T]
	notifyTimeout time.Duration
	logger        logger.Logger
}

var _ Publisher[string] = (*EventDispatcher[string])(nil)

func NewEventDispatcher[T any](name string, notifyTimeout time.Duration, log logger.Logger) *EventDispatcher[T] {
	res := &EventDispatcher[T]{
		name:          name,
		notifyTimeout: notifyTimeout,
		logger:        log.GetLogger(name),
		observers:     make(map[string]Observer[T]),
	}
	// check for correct timeout
	if res.notifyTimeout <= 0 {
		res.notifyTimeout = DefaultNotifyTimeout
	}

	return res
}

func (ed *EventDispatcher[T]) Register(observer Observer[T]) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	if utils.IsNil(observer) {
		return
	}

	ed.observers[observer.GetName()] = observer
}

func (ed *EventDispatcher[T]) Unregister(observer Observer[T]) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	if utils.IsNil(observer) {
		return
	}

	delete(ed.observers, observer.GetName())
}

func (ed *EventDispatcher[T]) Notify(ctx context.Context, data T) {
	ed.logger.Debugf("pub/sub event dispatcher %s Notify start", ed.GetName())
	defer ed.logger.Debugf("pub/sub event dispatcher %s Notify finish", ed.GetName())

	ed.mu.RLock()
	if len(ed.observers) == 0 {
		ed.mu.RUnlock()
		return
	}

	observers := make([]Observer[T], 0, len(ed.observers))
	for _, observer := range ed.observers {
		observers = append(observers, observer)
	}
	ed.mu.RUnlock()

	go ed.internalNotify(context.WithoutCancel(ctx), data, observers)
}

func (ed *EventDispatcher[T]) internalNotify(ctx context.Context, data T, observers []Observer[T]) {
	ed.logger.Debugf("pub/sub event dispatcher %s internalNotify start", ed.GetName())
	defer ed.logger.Debugf("pub/sub event dispatcher %s internalNotify finish", ed.GetName())

	var wg sync.WaitGroup
	asyncCtx, asyncCancel := context.WithTimeout(ctx, ed.notifyTimeout)
	defer asyncCancel()

	for _, observer := range observers {
		obs := observer
		wg.Add(1)
		go func(observe Observer[T]) {
			ed.logger.Debugf("pub/sub event dispatcher %s observer %s start", ed.GetName(), observe.GetName())
			defer ed.logger.Debugf("pub/sub event dispatcher %s observer %s finish", ed.GetName(), observe.GetName())

			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					// Превращаем панику в читаемую ошибку для логов
					var recoveryErr error
					if e, ok := r.(error); ok {
						recoveryErr = errs.NewCommonError("panic recovery", e)
					} else {
						recoveryErr = errs.NewCommonError(fmt.Sprintf("panic recovery [%v]", r), nil)
					}

					ed.logger.Errorf("pub/sub event dispatcher %s observer %s panic recovery %v", ed.GetName(), observe.GetName(), recoveryErr)
				}
			}()
			if err := observe.OnNotify(asyncCtx, data); err != nil {
				ed.logger.Errorf("pub/sub event dispatcher %s observer %s on notify got error %v", ed.GetName(), observe.GetName(), err)
			}
		}(obs)
	}
	wg.Wait()
}

func (ed *EventDispatcher[T]) GetName() string {
	return ed.name
}
