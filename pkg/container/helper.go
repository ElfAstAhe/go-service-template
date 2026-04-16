package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

func GetInstance[T any](container Container, name string) (T, error) {
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
		return errs.NewContainerValidateError("", "GetInstance", "container nil", nil)
	}
	if name == "" {
		return errs.NewContainerValidateError(container.GetName(), "GetInstance", "instance name is empty", nil)
	}

	return nil
}
