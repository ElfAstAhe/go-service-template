package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	InstanceTM              string = "TransactionManager"
	InstanceTestGetUC       string = "TestGetUC"
	InstanceTestGetByCodeUC string = "TestGetByCodeUC"
	InstanceTestListUC      string = "TestListUC"
	InstanceTestSaveUC      string = "TestSaveUC"
	InstanceTestDeleteUC    string = "TestDeleteUC"
)

type UseCaseContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*UseCaseContainer)(nil)
var _ container.LazyContainer = (*UseCaseContainer)(nil)

func NewUseCaseContainer(orchestrator container.Orchestrator) *UseCaseContainer {
	return &UseCaseContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(UseCaseContainerName, orchestrator),
	}
}

func (ucc *UseCaseContainer) Init(ctx context.Context) error {
	err := errors.Join(
		ucc.RegisterProvider(InstanceTM, ucc.providerTM),
		ucc.RegisterProvider(InstanceTestGetUC, ucc.providerTestGetUC),
		ucc.RegisterProvider(InstanceTestGetByCodeUC, ucc.providerTestGetByCodeUC),
		ucc.RegisterProvider(InstanceTestListUC, ucc.providerTestListUC),
		ucc.RegisterProvider(InstanceTestSaveUC, ucc.providerTestSaveUC),
		ucc.RegisterProvider(InstanceTestDeleteUC, ucc.providerTestDeleteUC),
	)
	if err != nil {
		return errs.NewContainerError(ucc.GetName(), "container init: register providers failed", err)
	}

	return nil
}
