package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/migration"
)

const (
	InstanceDB         string = "DB"
	InstanceDBMigrator string = "DBMigrator"
)

type PgContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*PgContainer)(nil)

func NewPgContainer(orchestrator container.Orchestrator) *PgContainer {
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
		pc.RegisterProvider(InstanceDBMigrator, pc.providerDBMigrator),
	)
	err := errors.Join(initErrs...)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: register providers failed", err)
	}
	// init instances
	db, err := container.GetInstance[*postgres.PgDB](pc, InstanceDB)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: init db failed", err)
	}
	// check db connection
	err = db.Ping(initCtx)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: check db failed", err)
	}
	// data migration
	migrator, err := container.GetInstance[migration.Migrator](pc, InstanceDBMigrator)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: init migrator failed", err)
	}
	err = migrator.Up(initCtx)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: up migrator failed", err)
	}

	return nil
}
