package container

import (
	"fmt"
	"io"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type BaseContainer struct {
	name      string
	mu        sync.RWMutex
	instances map[string]any
	log       logger.Logger
}

func NewBaseContainer(
	name string,
	log logger.Logger,
) *BaseContainer {
	return &BaseContainer{
		name:      name,
		instances: make(map[string]any),
		log:       log.GetLogger("BaseContainer"),
	}
}

func (bc *BaseContainer) GetName() string {
	return bc.name
}

func (bc *BaseContainer) Init() error {
	return errs.NewNotImplementedError(nil)
}

func (bc *BaseContainer) Close() error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	closeErrs := make([]error, 0)
	for i := len(bc.instances) - 1; i >= 0; i-- {
		if closer, ok := bc.instances[i].(io.Closer); ok {
			err := closer.Close()
			if err != nil {
				closeErrs = append(closeErrs, err)
				bc.GetLogger().Errorf(fmt.Sprintf("container [%s] close instance [%s] type [%s] failed", bc.GetName(), bc.instances[i], bc.name))
			}
		}
	}
}

func (bc *BaseContainer) Add(name string, instance any) error {
	if err := bc.commonValidate("BaseContainer.Add", name); err != nil {
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

func (bc *BaseContainer) Remove(name string) error {
	if err := bc.commonValidate("BaseContainer.Remove", name); err != nil {
		return err
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	if _, ok := bc.instances[name]; !ok {
		return errs.NewContainerError(bc.GetName(), fmt.Sprintf("instance %s not found", name), nil)
	}

	delete(bc.instances, name)

	return nil
}

func (bc *BaseContainer) GetInstance(name string) (any, error) {
	if err := bc.commonValidate("BaseContainer.GetInstance", name); err != nil {
		return nil, err
	}

	bc.mu.RLock()
	defer bc.mu.RUnlock()

	res, ok := bc.instances[name]
	if !ok {
		return nil, errs.NewContainerError(bc.GetName(), fmt.Sprintf("instance %s not found", name), nil)
	}

	return res, nil
}

func (bc *BaseContainer) AllInstances() map[string]any {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	res := make(map[string]any, len(bc.instances))
	for k, v := range bc.instances {
		res[k] = v
	}

	return res
}

func (bc *BaseContainer) GetLogger() logger.Logger {
	return bc.log
}

func (bc *BaseContainer) commonValidate(op string, name string) error {
	if name == "" {
		return errs.NewContainerValidateError(bc.GetName(), op, "name is empty", nil)
	}

	return nil
}
