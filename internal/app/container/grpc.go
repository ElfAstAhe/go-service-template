package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	InstanceGRPCService string = "grpcService"
	InstanceGRPCRunner  string = "grpcRunner"
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
	err := errors.Join(
		gc.RegisterProvider(InstanceGRPCService, gc.providerGRPCService),
		gc.RegisterProvider(InstanceGRPCRunner, gc.providerGRPCRunner),
	)
	if err != nil {
		return errs.NewContainerError(gc.GetName(), "container init: register providers failed", err)
	}

	return nil
}
