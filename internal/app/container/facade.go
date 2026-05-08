package container

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
)

type FacadeContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*FacadeContainer)(nil)
var _ container.LazyContainer = (*FacadeContainer)(nil)

func NewFacadeContainer(orchestrator container.Orchestrator) *FacadeContainer {
	return &FacadeContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(FacadeContainerName, orchestrator),
	}
}

func (fc *FacadeContainer) Init(ctx context.Context) error {
	return nil
}
