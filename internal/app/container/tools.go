package container

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
)

// ToolsContainer utils and helpers instances
type ToolsContainer struct {
	*container.BaseContainer
}

var _ container.Container = (*ToolsContainer)(nil)

func NewToolsContainer(orchestrator container.Orchestrator) *ToolsContainer {
	return &ToolsContainer{
		BaseContainer: container.NewBaseContainer(ToolsContainerName, orchestrator),
	}
}

func (tc *ToolsContainer) Init(ctx context.Context) error {
	return nil
}
