package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

const (
	InstanceHTTPRouter string = "HTTPRouter"
	InstanceHTTPRunner string = "HTTPRunner"
)

type HTTPContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*HTTPContainer)(nil)
var _ container.LazyContainer = (*HTTPContainer)(nil)

func NewHTTPContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *HTTPContainer {
	return &HTTPContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(
			container.WithLazyName(HTTPContainerName),
			container.WithLazyOrchestrator(orchestrator),
			container.WithLazyLogger(log),
		),
	}
}

func (hc *HTTPContainer) Init(initCtx context.Context) error {
	err := errors.Join(
		hc.RegisterProvider(InstanceHTTPRouter, hc.providerChiRouter),
		hc.RegisterProvider(InstanceHTTPRunner, hc.providerHTTPRunner),
	)
	if err != nil {
		return errs.NewContainerError(hc.GetName(), "container init: register providers failed", err)
	}

	return nil
}
