package container

import (
	"github.com/ElfAstAhe/go-service-template/internal/facade"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

func (fc *FacadeContainer) providerTestFacade() (any, error) {
	getUC, err := container.GetInstance[usecase.TestGetUseCase](InstanceTestGetUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	getByCodeUC, err := container.GetInstance[usecase.TestGetByCodeUseCase](InstanceTestGetByCodeUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	listUC, err := container.GetInstance[usecase.TestListUseCase](InstanceTestListUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	saveUC, err := container.GetInstance[usecase.TestSaveUseCase](InstanceTestSaveUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	deleteUC, err := container.GetInstance[usecase.TestDeleteUseCase](InstanceTestDeleteUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}

	return facade.NewTest(getUC, getByCodeUC, listUC, saveUC, deleteUC), nil
}

func (fc *FacadeContainer) providerTest2Facade() (any, error) {
	getUC, err := container.GetInstance[usecase.Test2GetUseCase](InstanceTest2GetUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	getByCodeUC, err := container.GetInstance[usecase.Test2GetByCodeUseCase](InstanceTest2GetByCodeUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	listUC, err := container.GetInstance[usecase.Test2ListUseCase](InstanceTest2ListUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	saveUC, err := container.GetInstance[usecase.Test2SaveUseCase](InstanceTest2SaveUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	deleteUC, err := container.GetInstance[usecase.Test2DeleteUseCase](InstanceTest2DeleteUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}

	return facade.NewTest2(getUC, getByCodeUC, listUC, saveUC, deleteUC), nil
}
