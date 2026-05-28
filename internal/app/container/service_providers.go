package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/hellofresh/health-go/v5"
)

func (sc *ServiceContainer) providerHealthStatus(name string) (any, error) {
	res, err := health.New(health.WithComponent(health.Component{
		Name:    config.AppName,
		Version: config.AppVersion,
	}))
	if err != nil {
		return nil, errs.NewContainerError(sc.GetName(), fmt.Sprintf("provider: create %s instance failed", name), err)
	}

	return res, nil
}
