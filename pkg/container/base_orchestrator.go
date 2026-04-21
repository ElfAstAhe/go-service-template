package container

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type BaseOrchestrator struct {
	mu       sync.RWMutex
	items    map[string]Container
	regOrder []string
	log      logger.Logger
}

func NewBaseOrchestrator(log logger.Logger) *BaseOrchestrator {
	return &BaseOrchestrator{
		log:      log.GetLogger("BaseOrchestrator"),
		items:    make(map[string]Container),
		regOrder: make([]string, 0),
	}
}

func (o *BaseOrchestrator) Init(ctx context.Context) error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	for _, name := range o.regOrder {
		ctn := o.items[name]
		o.log.Infof("initializing layer [%s]...", name)
		if err := ctn.Init(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (o *BaseOrchestrator) Close(ctx context.Context) error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// LIFO порядок: идем по слайсу имен с конца
	for i := len(o.regOrder) - 1; i >= 0; i-- {
		name := o.regOrder[i]
		if ctn, ok := o.items[name]; ok {
			o.log.Infof("closing layer [%s]...", name)
			if err := ctn.Close(ctx); err != nil {
				// Только логируем ошибку, продолжаем закрывать остальные
				o.log.Errorf("failed to close layer [%s]: %v", name, err)
			}
		}
	}
	return nil
}

func (o *BaseOrchestrator) Register(container Container) error {
	if err := o.validateContainer("BaseOrchestrator.Register", container); err != nil {
		return err
	}
	if o.HasContainer(container.GetName()) {
		return errs.NewContainerError("orchestrator", fmt.Sprintf("container [%s] already registered", container.GetName()), nil)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.items[container.GetName()] = container
	o.regOrder = append(o.regOrder, container.GetName())

	return nil
}

func (o *BaseOrchestrator) Unregister(name string) error {
	if err := o.validateName("BaseOrchestrator.Unregister", name); err != nil {
		return err
	}
	if !o.HasContainer(name) {
		return errs.NewContainerError("orchestrator", fmt.Sprintf("container [%s] not registered", name), nil)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.items, name)
	o.regOrder = slices.DeleteFunc(o.regOrder, func(item string) bool {
		return item == name
	})

	return nil
}

func (o *BaseOrchestrator) GetContainer(name string) (Container, error) {
	o.mu.RLock()
	res, ok := o.items[name]
	o.mu.RUnlock()
	if !ok {
		return nil, errs.NewContainerError("orchestrator", fmt.Sprintf("container [%s] not found", name), nil)
	}

	return res, nil
}

func (o *BaseOrchestrator) HasContainer(name string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	_, ok := o.items[name]

	return ok
}

// GetRunners — вспомогательный метод (пылесос)
func (o *BaseOrchestrator) GetRunners() []Runner {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var res []Runner
	for _, name := range o.regOrder {
		ctn := o.items[name]
		// Проходим по всем ГОТОВЫМ инстансам в контейнере
		for _, inst := range ctn.AllInstances() {
			if r, ok := inst.(Runner); ok {
				res = append(res, r)
			}
		}
	}
	return res
}

func (o *BaseOrchestrator) validateName(op, name string) error {
	if name == "" {
		return errs.NewContainerValidateError("orchestrator", op, "name is empty", nil)
	}

	return nil
}

func (o *BaseOrchestrator) validateContainer(op string, container Container) error {
	if utils.IsNil(container) {
		return errs.NewContainerValidateError("orchestrator", op, "container is nil", nil)
	}

	return nil
}
