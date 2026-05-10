package container

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
)

type GRPCContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*GRPCContainer)(nil)
var _ container.LazyContainer = (*GRPCContainer)(nil)

func NewGRPCContainer(orchestrator container.Orchestrator) *GRPCContainer {
	return &GRPCContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(GRPCContainerName, orchestrator),
	}
}

func (gc *GRPCContainer) Init(initCtx context.Context) error {
	return nil
}
