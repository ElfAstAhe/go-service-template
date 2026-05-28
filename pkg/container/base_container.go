package container

import (
	"context"
	"fmt"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

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
	return nil
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
		return nil, errs.NewContainerError(bc.GetName(), fmt.Sprintf("instance %s is not registered", name), nil)
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

func (bc *BaseContainer) GetOrchestrator() Orchestrator {
	return bc.orchestrator
}
