package container

import (
	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/facade"
	"github.com/ElfAstAhe/go-service-template/internal/transport/rest"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/http"
	"github.com/hellofresh/health-go/v5"
)

func (hc *HTTPContainer) providerChiRouter(name string) (any, error) {
	appCnt, err := hc.GetOrchestrator().GetContainer(AppContainerName)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve container failed", err)
	}
	confInst, err := container.GetInstance[*config.Config](appCnt, InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	logInst, err := container.GetInstance[logger.Logger](appCnt, InstanceLogger)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	readyz, err := container.GetInstance[http.ReadyzFunc](appCnt, InstanceApplicationReady)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	facadeCnt, err := hc.GetOrchestrator().GetContainer(FacadeContainerName)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve container failed", err)
	}
	testFacadeInst, err := container.GetInstance[facade.TestFacade](facadeCnt, InstanceTestFacade)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve instance failed", err)
	}
	srvCnt, err := hc.GetOrchestrator().GetContainer(ServiceContainerName)
	if err != nil {
		return nil, errs.NewContainerError(hc.GetName(), "provider: retrieve container failed", err)
	}
	healthInst, err := container.GetInstance[*health.Health](srvCnt, InstanceHealthStatus)
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

func (hc *HTTPContainer) providerHTTPRunner(name string) (any, error) {
	// ToDo: implement

	return nil, nil
}
