package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/repository"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	ContainerName    string = "postgres"
	InstanceDB       string = "AppDB"
	InstanceTestRepo string = "TestRepo"
)

type PgContainer struct {
	*container.BaseLazyContainer
	conf *config.Config
}

var _ container.Container = (*PgContainer)(nil)

func NewContainer(conf *config.Config) *PgContainer {
	res := &PgContainer{
		conf: conf,
	}
	res.BaseLazyContainer = container.NewBaseLazyContainer(ContainerName)

	return res
}

func (pc *PgContainer) Init(initCtx context.Context) error {
	// add providers
	initErrs := make([]error, 0)
	initErrs = append(initErrs,
		pc.RegisterProvider(InstanceDB, pc.providerDB),
		pc.RegisterProvider(InstanceTestRepo, pc.providerTestRepository),
	)
	err := errors.Join(initErrs...)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: register providers failed", err)
	}
	// init instances
	db, err := container.GetInstance[*PgDB](pc, "AppDB")
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
	res, err := NewPgDB(pc.conf.DB)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), fmt.Sprintf("create %s instance failed", name), err)
	}

	return res, nil
}

func (pc *PgContainer) providerTestRepository(name string) (any, error) {
	database, err := container.GetInstance[*PgDB](pc, InstanceDB)
	if err != nil {
		return nil, err
	}
	res, err := NewTestRepository(database, database)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), "create test repo instance failed", err)
	}

	return repository.NewTestMetricsRepository(res), nil
}
