package container

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// BaseContainer simple instance storage
type BaseContainer struct {
	name         string
	mu           sync.RWMutex
	instances    map[string]any
	orchestrator Orchestrator
	logger       logger.Logger
}

func NewBaseContainer(
	opts ...Option,
) *BaseContainer {
	options := &Options{}

	for _, o := range opts {
		o(options)
	}

	return &BaseContainer{
		name:         options.Name,
		instances:    make(map[string]any),
		orchestrator: options.Orchestrator,
		logger:       options.Logger.GetLogger("BaseContainer"),
	}
}

var _ Container = (*BaseContainer)(nil)

func (bc *BaseContainer) GetName() string {
	return bc.name
}

func (bc *BaseContainer) Init(initCtx context.Context) error {
	return errs.NewNotImplementedError(nil)
}

//goland:noinspection DuplicatedCode
func (bc *BaseContainer) Close(closeCtx context.Context) error {
	bc.logger.Debugf("container %s close: started", bc.GetName())
	defer bc.logger.Debugf("container %s close: finished", bc.GetName())

	var wg sync.WaitGroup

	// 1. Быстро копируем ссылки на инстансы под RLock
	bc.mu.RLock()
	toClosing := make([]*closeInstance, 0, len(bc.instances))
	for name, instance := range bc.instances {
		toClosing = append(toClosing, newCloseInstance(name, instance))
	}
	bc.mu.RUnlock()

	bc.logger.Debugf("container %s close: got %d instances to close", bc.GetName(), len(toClosing))

	// 2. Полностью очищаем мапу под Lock, так как контейнер уничтожается
	bc.mu.Lock()
	bc.instances = make(map[string]any)
	bc.mu.Unlock()

	closeChan := make(chan struct{})
	closeErrs := utils.NewConcurrentList[error]()
	for _, toClose := range toClosing {
		if inst, ok := toClose.Instance.(SimpleCloser); ok {
			wg.Add(1)
			go func(name string, closer SimpleCloser) {
				bc.logger.Debugf("container %s close: simple closer for instance %s start", bc.GetName(), name)
				defer bc.logger.Debugf("container %s close: simple closer for instance %s finish", bc.GetName(), name)

				defer wg.Done()

				if err := closer.Close(); err != nil {
					closeErrs.Append(err)
					bc.logger.Debugf("container %s close: simple closer for instance %s failed: %v", bc.GetName(), name, err)
				} else {
					bc.logger.Debugf("container %s close: simple closer for instance %s done", bc.GetName(), name)
				}
			}(toClose.Name, inst)
		} else if inst, ok := toClose.Instance.(ContextCloser); ok {
			wg.Add(1)
			go func(name string, closer ContextCloser) {
				bc.logger.Debugf("container %s close: context closer for instance %s start", bc.GetName(), name)
				defer bc.logger.Debugf("container %s close: context closer for instance %s finish", bc.GetName(), name)

				defer wg.Done()

				if err := closer.Close(closeCtx); err != nil {
					closeErrs.Append(err)
					bc.logger.Debugf("container %s close: context closer for instance %s failed: %v", bc.GetName(), name, err)
				} else {
					bc.logger.Debugf("container %s close: context closer for instance %s done", bc.GetName(), name)
				}
			}(toClose.Name, inst)
		} else {
			bc.logger.Debugf("container %s close: no closer methods for instance %s", bc.GetName(), toClose.Name)
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
	bc.logger.Debugf("container %s register instance: started", bc.GetName())
	defer bc.logger.Debugf("container %s register instance: finished", bc.GetName())

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
	bc.logger.Debugf("container %s unregister instance: started", bc.GetName())
	defer bc.logger.Debugf("container %s unregister instance: finished", bc.GetName())

	if err := bc.Validate("BaseContainer.UnregisterInstance", name); err != nil {
		return err
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	delete(bc.instances, name)

	return nil
}

func (bc *BaseContainer) GetInstance(name string) (any, error) {
	bc.logger.Debugf("container %s get instance: started", bc.GetName())
	defer bc.logger.Debugf("container %s get instance: finished", bc.GetName())

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
	bc.logger.Debugf("container %s all names: started", bc.GetName())
	defer bc.logger.Debugf("container %s all names: finished", bc.GetName())

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
	bc.logger.Debugf("container %s validate: started", bc.GetName())
	defer bc.logger.Debugf("container %s validate: finished", bc.GetName())

	if name == "" {
		return errs.NewContainerValidateError(bc.GetName(), op, "name is empty", nil)
	}

	return nil
}

func (bc *BaseContainer) IsRegistered(name string) bool {
	bc.logger.Debugf("container %s is registered: started", bc.GetName())
	defer bc.logger.Debugf("container %s is registered: finished", bc.GetName())

	bc.mu.RLock()
	defer bc.mu.RUnlock()

	_, ok := bc.instances[name]

	return ok
}

func (bc *BaseContainer) HasInstance(name string) bool {
	bc.logger.Debugf("container %s has instance: started", bc.GetName())
	defer bc.logger.Debugf("container %s has instance: finished", bc.GetName())

	bc.mu.RLock()
	defer bc.mu.RUnlock()

	_, ok := bc.instances[name]

	return ok
}

func (bc *BaseContainer) GetOrchestrator() Orchestrator {
	return bc.orchestrator
}
