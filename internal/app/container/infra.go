package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/metrics"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type InfraContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*InfraContainer)(nil)
var _ container.LazyContainer = (*InfraContainer)(nil)

func NewInfraContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *InfraContainer {
	return &InfraContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(
			container.WithLazyName(InfraContainerName),
			container.WithLazyOrchestrator(orchestrator),
			container.WithLazyLogger(log),
		),
	}
}

//goland:noinspection GoUnusedParameter
func (ic *InfraContainer) Init(ctx context.Context) error {
	err := errors.Join()
	if err != nil {
		return errs.NewContainerError(ic.GetName(), "container init: register providers failed", err)
	}

	// setup metrics
	metrics.InitRepositoryMetrics()
	metrics.InitHTTPMetrics()
	InitGRPCMetrics()

	return nil
}
