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

// PgContainer database connection and data migrations
type PgContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*PgContainer)(nil)
var _ container.LazyContainer = (*PgContainer)(nil)

func NewPgContainer(orchestrator container.Orchestrator) *PgContainer {
	res := &PgContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(DBContainerName, orchestrator),
	}

	return res
}

func (pc *PgContainer) Init(initCtx context.Context) error {
	// add providers
	err := errors.Join(
		pc.RegisterProvider(InstanceDB, pc.providerDB),
		pc.RegisterProvider(InstanceDBMigrator, pc.providerDBMigrator),
	)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: register providers failed", err)
	}
	// init db instance
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
	// migrate up
	err = migrator.Up(initCtx)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: up migrator failed", err)
	}

	return nil
}
