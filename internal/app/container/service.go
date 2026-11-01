package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type ServiceContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*ServiceContainer)(nil)
var _ container.LazyContainer = (*ServiceContainer)(nil)

func NewServiceContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *ServiceContainer {
	return &ServiceContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(
			container.WithLazyName(ServiceContainerName),
			container.WithLazyOrchestrator(orchestrator),
			container.WithLazyLogger(log),
		),
	}
}

func (sc *ServiceContainer) Init(initCtx context.Context) error {
	// add providers
	err := errors.Join()
	if err != nil {
		return errs.NewContainerError(sc.GetName(), "container init: register providers failed", err)
	}

	return nil
}
