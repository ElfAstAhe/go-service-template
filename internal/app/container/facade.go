package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

const (
	InstanceTestFacade string = "TestFacade"
)

type FacadeContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*FacadeContainer)(nil)
var _ container.LazyContainer = (*FacadeContainer)(nil)

func NewFacadeContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *FacadeContainer {
	return &FacadeContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(
			container.WithLazyName(FacadeContainerName),
			container.WithLazyOrchestrator(orchestrator),
			container.WithLazyLogger(log),
		),
	}
}

func (fc *FacadeContainer) Init(ctx context.Context) error {
	err := errors.Join(
		fc.RegisterProvider(InstanceTestFacade, fc.providerTestFacade),
	)
	if err != nil {
		return errs.NewContainerError(fc.GetName(), "container init: register providers failed", err)
	}

	return nil
}
