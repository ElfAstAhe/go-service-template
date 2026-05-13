package container

import (
	"github.com/ElfAstAhe/go-service-template/internal/facade"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

func (fc *FacadeContainer) providerTestFacade(name string) (any, error) {
	ucCnt, err := fc.GetOrchestrator().GetContainer(UseCaseContainerName)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve container failed", err)
	}
	getUC, err := container.GetInstance[usecase.TestGetUseCase](ucCnt, InstanceTestGetUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	getByCodeUC, err := container.GetInstance[usecase.TestGetByCodeUseCase](ucCnt, InstanceTestGetByCodeUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	listUC, err := container.GetInstance[usecase.TestListUseCase](ucCnt, InstanceTestListUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	saveUC, err := container.GetInstance[usecase.TestSaveUseCase](ucCnt, InstanceTestSaveUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}
	deleteUC, err := container.GetInstance[usecase.TestDeleteUseCase](ucCnt, InstanceTestDeleteUC)
	if err != nil {
		return nil, errs.NewContainerError(fc.GetName(), "provider: retrieve instance failed", err)
	}

	return facade.NewTestFacade(getUC, getByCodeUC, listUC, saveUC, deleteUC), nil
}
