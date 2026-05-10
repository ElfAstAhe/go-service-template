package container

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
)

type HTTPContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*HTTPContainer)(nil)
var _ container.LazyContainer = (*HTTPContainer)(nil)

func NewHTTPContainer(orchestrator container.Orchestrator) *HTTPContainer {
	return &HTTPContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(HTTPContainerName, orchestrator),
	}
}

func (hc *HTTPContainer) Init(initCtx context.Context) error {
	return nil
}
