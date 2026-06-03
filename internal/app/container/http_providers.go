package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/facade"
	"github.com/ElfAstAhe/go-service-template/internal/transport/rest"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/hellofresh/health-go/v5"
)

//goland:noinspection DuplicatedCode
func (hc *HTTPContainer) providerChiRouter() (any, error) {
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	logInst, err := container.GetInstance[logger.Logger](InstanceLogger)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	readyz, err := container.GetInstance[func() bool](InstanceApplicationReady)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	testFacadeInst, err := container.GetInstance[facade.TestFacade](InstanceTestFacade)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	healthInst, err := container.GetInstance[*health.Health](InstanceHealthStatus)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}

	return rest.NewAppChiRouter(
		confInst.HTTP,
		confInst.Telemetry,
		logInst,
		healthInst,
		nil,
		readyz,
		testFacadeInst,
	), nil
}

//goland:noinspection DuplicatedCode
func (hc *HTTPContainer) providerHTTPRunner() (any, error) {
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	logInst, err := container.GetInstance[logger.Logger](InstanceLogger)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	routerInst, err := container.GetInstance[http.Router](InstanceHTTPRouter)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}

	runner, err := http.NewRunner(
		http.WithName("main-http-server"),
		http.WithConfig(confInst.HTTP),
		http.WithLogger("http_server", logInst),
		http.WithRouter(routerInst),
	)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), fmt.Sprintf("provider: create %s failed", InstanceHTTPRunner), err)
	}

	return runner, nil
}
