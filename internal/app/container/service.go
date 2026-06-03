package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/hellofresh/health-go/v5"
	healthPg "github.com/hellofresh/health-go/v5/checks/pgx5"
)

const (
	InstanceHealthStatus string = "health-status"
)

type ServiceContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*ServiceContainer)(nil)
var _ container.LazyContainer = (*ServiceContainer)(nil)

func NewServiceContainer(orchestrator container.Orchestrator) *ServiceContainer {
	return &ServiceContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(ServiceContainerName, orchestrator),
	}
}

func (sc *ServiceContainer) Init(initCtx context.Context) error {
	// add providers
	err := errors.Join(
		sc.RegisterProvider(InstanceHealthStatus, sc.providerHealthStatus),
	)
	if err != nil {
		return errs.NewContainerError(sc.GetName(), "container init: register providers failed", err)
	}
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return errs.NewContainerError(sc.GetName(), "container init: retrieve instance failed", err)
	}
	// init health checks
	dbInst, err := container.GetInstance[db.DB](InstanceDB)
	if err != nil {
		return errs.NewContainerError(sc.GetName(), "container init: retrieve instance failed", err)
	}
	healthStatus, err := container.GetInstance[*health.Health](InstanceHealthStatus)
	if err != nil {
		return errs.NewContainerError(sc.GetName(), "container init: retrieve instance failed", err)
	}
	// Регистрируем Postgres. Либа сама будет делать Ping
	err = healthStatus.Register(health.Config{
		Name:      dbInst.GetDriver(),
		Timeout:   confInst.DB.ConnTimeout,
		SkipOnErr: false,
		Check: healthPg.New(healthPg.Config{
			DSN: confInst.DB.DSN,
		}),
	})
	if err != nil {
		return errs.NewContainerError(sc.GetName(), "container init: register health status check failed", err)
	}

	return nil
}
