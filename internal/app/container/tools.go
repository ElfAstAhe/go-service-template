package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// ToolsContainer utils and helpers instances
type ToolsContainer struct {
	*container.BaseContainer
}

var _ container.Container = (*ToolsContainer)(nil)

func NewToolsContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *ToolsContainer {
	return &ToolsContainer{
		BaseContainer: container.NewBaseContainer(
			container.WithName(ToolsContainerName),
			container.WithOrchestrator(orchestrator),
			container.WithLogger(log),
		),
	}
}

func (tc *ToolsContainer) Init(ctx context.Context) error {
	// add providers
	err := errors.Join()
	if err != nil {
		return errs.NewContainerError(tc.GetName(), "container init: register providers failed", err)
	}

	return nil
}
