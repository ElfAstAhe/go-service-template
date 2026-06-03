package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

var defaultOrchestrator Orchestrator

func SetDefaultOrchestrator(orc Orchestrator) {
	defaultOrchestrator = orc
}

func GetInstance[T any](name string) (T, error) {
	var nilRes T
	if utils.IsNil(defaultOrchestrator) {
		return nilRes, errs.NewContainerValidateError("default orchestrator", "GetInstance", "default orchestrator not set up", nil)
	}

	var err error
	var res T
	for _, cnt := range defaultOrchestrator.AllContainers() {
		if cnt.IsRegistered(name) {
			// look for first equal name and non nil, so use unique naming for instances
			res, err = GetContainerInstance[T](cnt, name)
			if err != nil {
				return nilRes, err
			}
			if !utils.IsNil(res) {
				return res, nil
			}
		}
	}

	return nilRes, errs.NewContainerNotFoundError(fmt.Sprintf("instance %s not found in all registered containers", name), nil)
}

func GetContainerInstance[T any](container Container, name string) (T, error) {
	var nilRes T
	// validate
	if err := getInstanceValidate(container, name); err != nil {
		return nilRes, err
	}
	// retrieve instance
	instance, err := container.GetInstance(name)
	if err != nil {
		return nilRes, err
	}
	// check nil
	if utils.IsNil(instance) {
		return nilRes, nil
	}
	// transform
	res, ok := instance.(T)
	if !ok {
		return nilRes, errs.NewContainerError(container.GetName(), fmt.Sprintf("instance type [%s] mismatch", utils.GetFullTypeName(instance)), nil)
	}

	return res, nil
}

func getInstanceValidate(container Container, name string) error {
	if container == nil {
		return errs.NewContainerValidateError("container", "GetContainerInstance", "container nil", nil)
	}
	if name == "" {
		return errs.NewContainerValidateError(container.GetName(), "GetContainerInstance", "instance name is empty", nil)
	}

	return nil
}
