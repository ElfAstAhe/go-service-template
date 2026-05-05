package container

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	InstanceDB string = "AppDB"
)

type PgContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*PgContainer)(nil)

func NewContainer(orchestrator container.Orchestrator) *PgContainer {
	res := &PgContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(DBContainerName, orchestrator),
	}

	return res
}

func (pc *PgContainer) Init(initCtx context.Context) error {
	// add providers
	initErrs := make([]error, 0)
	initErrs = append(initErrs,
		pc.RegisterProvider(InstanceDB, pc.providerDB),
	)
	err := errors.Join(initErrs...)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: register providers failed", err)
	}
	// init instances
	db, err := container.GetInstance[*postgres.PgDB](pc, InstanceDB)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: init instances failed", err)
	}

	err = db.Ping(initCtx)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "init db failed", err)
	}

	return nil
}

func (pc *PgContainer) providerDB(name string) (any, error) {
	appCnt, err := pc.GetOrchestrator().GetContainer(AppContainerName)
	if err != nil {
		return nil, err
	}
	conf, err := container.GetInstance[*config.Config](appCnt, ConfigInstance)
	if err != nil {
		return nil, err
	}
	res, err := postgres.NewPgDB(conf.DB)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), fmt.Sprintf("create %s instance failed", name), err)
	}

	return res, nil
}
func (pc *PgContainer) providerDBMigrator(name string) (any, error) {

}
