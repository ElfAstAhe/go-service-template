package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/repository"
	"github.com/ElfAstAhe/go-service-template/internal/repository/metrics"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	"github.com/ElfAstAhe/go-service-template/internal/repository/trace"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

func (rc *RepositoryContainer) providerTestRepository() (any, error) {
	dbInst, err := container.GetInstance[db.DB](InstanceDB)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), "provider: retrieve instance failed", err)
	}
	res, err := postgres.NewTestRepository(dbInst, dbInst)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), fmt.Sprintf("provider: create [%s] repo instance failed", InstanceTestRepo), err)
	}

	return res, nil
}

func (rc *RepositoryContainer) providerTest2Repository() (any, error) {
	dbInst, err := container.GetInstance[db.DB](InstanceDB)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), "provider: retrieve instance failed", err)
	}
	res, err := repository.NewTest2(dbInst.GetDriver(), dbInst, dbInst)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), fmt.Sprintf("provider: create [%s] repo instance failed", InstanceTestRepo), err)
	}

	return res, nil
}

func (rc *RepositoryContainer) providerTestMetrics() (any, error) {
	repoInst, err := container.GetInstance[domain.TestRepository](InstanceTestRepo)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), "provider: retrieve instance failed", err)
	}

	return metrics.NewTestRepository(repoInst), nil
}

func (rc *RepositoryContainer) providerTest2Metrics() (any, error) {
	repoInst, err := container.GetInstance[domain.Test2Repository](InstanceTest2Repo)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), "provider: retrieve instance failed", err)
	}

	return metrics.NewTest2Repository(repoInst), nil
}

func (rc *RepositoryContainer) providerTestTrace() (any, error) {
	metricsInst, err := container.GetInstance[domain.TestRepository](InstanceTestMetrics)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), "provider: retrieve instance failed", err)
	}

	return trace.NewTestRepository(metricsInst), nil
}

func (rc *RepositoryContainer) providerTest2Trace() (any, error) {
	metricsInst, err := container.GetInstance[domain.Test2Repository](InstanceTest2Metrics)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), "provider: retrieve instance failed", err)
	}
	res, err := trace.NewTest2Repository(metricsInst)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), fmt.Sprintf("provider: create [%s] repo instance failed", InstanceTest2Trace), err)
	}

	return res, nil
}
