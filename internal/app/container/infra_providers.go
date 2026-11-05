package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/hellofresh/health-go/v5"
)

func (ic *InfraContainer) providerHealthMonitoring() (any, error) {
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(ic.GetName(), "provider: retrieve instance failed", err)
	}
	res, err := health.New(
		health.WithSystemInfo(),
		health.WithComponent(health.Component{
			Name:    confInst.App.NodeName,
			Version: config.AppVersion,
		}),
	)
	if err != nil {
		return nil, errs.NewContainerError(ic.GetName(), fmt.Sprintf("provider: initialize %s instance failed", InstanceHealthMonitoring), err)
	}

	return res, nil
}
