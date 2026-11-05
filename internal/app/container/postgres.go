package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/migration"
	"github.com/hellofresh/health-go/v5"
	healthPg "github.com/hellofresh/health-go/v5/checks/pgx5"
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
		return errs.NewContainerError(pc.GetName(), "provider: retrieve instance failed", err)
	}
	// check db connection
	err = dbInst.Ping(initCtx)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: check db failed", err)
	}
	// data migration
	migratorInst, err := container.GetInstance[migration.Migrator](InstanceDBMigrator)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "provider: retrieve instance failed", err)
	}
	// migrate up
	err = migratorInst.Up(initCtx)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: up migrator failed", err)
	}
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "provider: retrieve instance failed", err)
	}
	// db metrics
	healthInst, err := container.GetInstance[*health.Health](InstanceHealthMonitoring)
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "provider: retrieve instance failed", err)
	}
	err = healthInst.Register(health.Config{
		Name:      dbInst.GetDriver(),
		Timeout:   confInst.DB.ConnTimeout,
		SkipOnErr: true,
		Check: healthPg.New(healthPg.Config{
			DSN: confInst.DB.DSN,
		}),
	})
	if err != nil {
		return errs.NewContainerError(pc.GetName(), "container init: register db health monitoring failed", err)
	}

	return nil
}
