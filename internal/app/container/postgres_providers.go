package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	_ "github.com/ElfAstAhe/go-service-template/migrations/example-service"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/migration/goose"
)

func (pc *PgContainer) providerDB() (any, error) {
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), "provider: retrieve instance failed", err)
	}
	res, err := postgres.NewPgDB(confInst.DB)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), fmt.Sprintf("provider: create %s instance failed", InstanceDB), err)
	}

	return res, nil
}

func (pc *PgContainer) providerDBMigrator() (any, error) {
	logInst, err := container.GetInstance[logger.Logger](InstanceLogger)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), "provider: retrieve instance failed", err)
	}
	dbInst, err := container.GetInstance[*postgres.PgDB](InstanceDB)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), "provider: retrieve instance failed", err)
	}
	res, err := goose.NewDBMigrator(dbInst, logInst)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), fmt.Sprintf("provider: create %s instance failed", InstanceDBMigrator), err)
	}
	err = res.Initialize()
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), fmt.Sprintf("provider: initialize %s instance failed", InstanceDBMigrator), err)
	}

	return res, nil
}
