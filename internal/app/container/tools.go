package container

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
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
	return nil
}
