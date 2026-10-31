package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

const (
	InstanceUnitOfWork       string = "unit-of-work"
	InstanceTestGetUC        string = "TestGetUC"
	InstanceTestGetByCodeUC  string = "TestGetByCodeUC"
	InstanceTestListUC       string = "TestListUC"
	InstanceTestSaveUC       string = "TestSaveUC"
	InstanceTestDeleteUC     string = "TestDeleteUC"
	InstanceTest2GetUC       string = "test2-get-use-case-uc"
	InstanceTest2GetByCodeUC string = "test2-get-by-code-uc"
	InstanceTest2ListUC      string = "test2-list-uc"
	InstanceTest2SaveUC      string = "test2-save-uc"
	InstanceTest2DeleteUC    string = "test2-delete-uc"
)

type UseCaseContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*UseCaseContainer)(nil)
var _ container.LazyContainer = (*UseCaseContainer)(nil)

func NewUseCaseContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *UseCaseContainer {
	return &UseCaseContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(
			container.WithLazyName(UseCaseContainerName),
			container.WithLazyOrchestrator(orchestrator),
			container.WithLazyLogger(log),
		),
	}
}

func (ucc *UseCaseContainer) Init(ctx context.Context) error {
	err := errors.Join(
		ucc.RegisterProvider(InstanceUnitOfWork, ucc.providerUnitOfWork),
		ucc.RegisterProvider(InstanceTestGetUC, ucc.providerTestGetUC),
		ucc.RegisterProvider(InstanceTestGetByCodeUC, ucc.providerTestGetByCodeUC),
		ucc.RegisterProvider(InstanceTestListUC, ucc.providerTestListUC),
		ucc.RegisterProvider(InstanceTestSaveUC, ucc.providerTestSaveUC),
		ucc.RegisterProvider(InstanceTestDeleteUC, ucc.providerTestDeleteUC),
		ucc.RegisterProvider(InstanceTest2GetUC, ucc.providerTest2GetUC),
		ucc.RegisterProvider(InstanceTest2GetByCodeUC, ucc.providerTest2GetByCodeUC),
		ucc.RegisterProvider(InstanceTest2ListUC, ucc.providerTest2ListUC),
		ucc.RegisterProvider(InstanceTest2SaveUC, ucc.providerTest2SaveUC),
		ucc.RegisterProvider(InstanceTest2DeleteUC, ucc.providerTest2DeleteUC),
	)
	if err != nil {
		return errs.NewContainerError(ucc.GetName(), "container init: register providers failed", err)
	}

	return nil
}
