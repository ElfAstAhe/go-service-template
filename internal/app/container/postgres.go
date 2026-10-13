package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/migration"
)

const (
	InstanceDB         string = "DB"
	InstanceDBMigrator string = "DBMigrator"
	InstanceTM         string = "TransactionManager"
	InstanceTMOpts     string = "TransactionManagerOpts"
)

// PgContainer database connection and data migrations
type PgContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*PgContainer)(nil)
var _ container.LazyContainer = (*PgContainer)(nil)

func NewPgContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *PgContainer {
	return &PgContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(
			container.WithLazyName(DBContainerName),
			container.WithLazyOrchestrator(orchestrator),
			container.WithLazyLogger(log),
		),
	}
}

func (pc *PgContainer) Init(initCtx context.Context) error {
	// add providers
	err := errors.Join(
		pc.RegisterProvider(InstanceDB, pc.providerDB),
		pc.RegisterProvider(InstanceDBMigrator, pc.providerDBMigrator),
		pc.RegisterProvider(InstanceTM, pc.providerTM),
		pc.RegisterProvider(InstanceTMOpts, pc.providerTMOpts),
	)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: register providers failed", err)
	}
	// init db instance
	dbInst, err := container.GetInstance[db.DB](InstanceDB)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: init db failed", err)
	}
	// check db connection
	err = dbInst.Ping(initCtx)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: check db failed", err)
	}
	// data migration
	migratorInst, err := container.GetInstance[migration.Migrator](InstanceDBMigrator)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: init migrator failed", err)
	}
	// migrate up
	err = migratorInst.Up(initCtx)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: up migrator failed", err)
	}

	return nil
}
