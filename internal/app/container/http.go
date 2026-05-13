package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
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

func NewHTTPContainer(orchestrator container.Orchestrator) *HTTPContainer {
	return &HTTPContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(HTTPContainerName, orchestrator),
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
