package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	migrations "github.com/ElfAstAhe/go-service-template/pkg/migration/goose"
)

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
		return nil, errs.NewContainerError(pc.GetName(), fmt.Sprintf("provider: create %s instance failed", name), err)
	}

	return res, nil
}

func (pc *PgContainer) providerDBMigrator(name string) (any, error) {
	appCnt, err := pc.GetOrchestrator().GetContainer(AppContainerName)
	if err != nil {
		return nil, err
	}
	log, err := container.GetInstance[logger.Logger](appCnt, LoggerInstance)
	if err != nil {
		return nil, err
	}
	db, err := container.GetInstance[*postgres.PgDB](pc, name)
	if err != nil {
		return nil, err
	}
	res, err := migrations.NewDBMigrator(db, log)
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), fmt.Sprintf("provider: create %s instance failed", name), err)
	}
	err = res.Initialize()
	if err != nil {
		return nil, errs.NewContainerError(pc.GetName(), fmt.Sprintf("provider: initialize %s instance failed", name), err)
	}

	return res, nil
}
