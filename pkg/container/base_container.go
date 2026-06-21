package container

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type SimpleCloser interface {
	Close() error
}

type ContextCloser interface {
	Close(ctx context.Context) error
}

type BaseContainer struct {
	name         string
	mu           sync.RWMutex
	instances    map[string]any
	orchestrator Orchestrator
}

func NewBaseContainer(
	name string,
	orchestrator Orchestrator,
) *BaseContainer {
	return &BaseContainer{
		name:         name,
		instances:    make(map[string]any),
		orchestrator: orchestrator,
	}
}

var _ Container = (*BaseContainer)(nil)

func (bc *BaseContainer) GetName() string {
	return bc.name
}

func (bc *BaseContainer) Init(initCtx context.Context) error {
	return errs.NewNotImplementedError(nil)
}

func (bc *BaseContainer) Close(closeCtx context.Context) error {
	var wg sync.WaitGroup

	// 1. Быстро копируем ссылки на инстансы под RLock
	bc.mu.RLock()
	closables := make([]any, 0, len(bc.instances))
	for _, instance := range bc.instances {
		closables = append(closables, instance)
	}
	bc.mu.RUnlock()

	// 2. Полностью очищаем мапу под Lock, так как контейнер уничтожается
	bc.mu.Lock()
	bc.instances = make(map[string]any)
	bc.mu.Unlock()

	closeChan := make(chan struct{})
	closeErrs := utils.NewConcurrentList[error]()
	for _, instance := range closables {
		if inst, ok := instance.(SimpleCloser); ok {
			wg.Add(1)
			go func(closer SimpleCloser) {
				defer wg.Done()
				if err := closer.Close(); err != nil {
					closeErrs.Append(err)
				}
			}(inst)
		} else if inst, ok := instance.(ContextCloser); ok {
			wg.Add(1)
			go func(closer ContextCloser) {
				defer wg.Done()
				if err := closer.Close(closeCtx); err != nil {
					closeErrs.Append(err)
				}
			}(inst)
		}
	}
	go func() {
		defer close(closeChan)
		wg.Wait()
	}()
	select {
	case <-closeChan:
		if closeErrs.Len() > 0 {
			return errs.NewContainerError(bc.GetName(), "container close: close fails", errors.Join(closeErrs.Snapshot()...))
		}

		return nil
	case <-closeCtx.Done():
		return errs.NewContainerError(bc.GetName(), "container close: close timeout limit reached", nil)
	}
}

func (bc *BaseContainer) RegisterInstance(name string, instance any) error {
	if err := bc.Validate("BaseContainer.RegisterInstance", name); err != nil {
		return err
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	// check existence
	if _, ok := bc.instances[name]; ok {
		return errs.NewContainerError(bc.GetName(), fmt.Sprintf("instance %s already exists", name), nil)
	}

	bc.instances[name] = instance

	return nil
}

func (bc *BaseContainer) UnregisterInstance(name string) error {
	if err := bc.Validate("BaseContainer.UnregisterInstance", name); err != nil {
		return err
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	delete(bc.instances, name)

	return nil
}

func (bc *BaseContainer) GetInstance(name string) (any, error) {
	// validate
	if err := bc.Validate("BaseContainer.GetInstance", name); err != nil {
		return nil, err
	}

	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// retrieve instance
	res, ok := bc.instances[name]
	if !ok {
		return nil, errs.NewContainerNotFoundError(fmt.Sprintf("container [%s] instance [%s] is not registered ", bc.GetName(), name), nil)
	}

	return res, nil
}

func (bc *BaseContainer) AllNames() []string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if len(bc.instances) == 0 {
		return nil
	}

	// make a map copy
	res := make([]string, 0, len(bc.instances))
	for key, _ := range bc.instances {
		res = append(res, key)
	}

	return res
}

func (bc *BaseContainer) Validate(op string, name string) error {
	if name == "" {
		return errs.NewContainerValidateError(bc.GetName(), op, "name is empty", nil)
	}

	return nil
}

func (bc *BaseContainer) IsRegistered(name string) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	_, ok := bc.instances[name]

	return ok
}

func (bc *BaseContainer) HasInstance(name string) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	_, ok := bc.instances[name]

	return ok
}

func (bc *BaseContainer) GetOrchestrator() Orchestrator {
	return bc.orchestrator
}
