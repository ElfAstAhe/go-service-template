package container

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/repository"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	InstanceTestRepo string = "testRepo"
)

type RepositoryContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*RepositoryContainer)(nil)
var _ container.LazyContainer = (*RepositoryContainer)(nil)

func NewRepositoryContainer(orchestrator container.Orchestrator) *RepositoryContainer {
	return &RepositoryContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(RepositoryContainerName, orchestrator),
	}
}

func (rc *RepositoryContainer) Init(ctx context.Context) error {
	err := errors.Join(
		rc.RegisterProvider(InstanceTestRepo, rc.providerTestRepository),
	)
	if err != nil {
		return errs.NewContainerError(rc.GetName(), "container init: register providers failed", err)
	}

	return nil
}

func (rc *RepositoryContainer) providerTestRepository(name string) (any, error) {
	dbCnt, err := rc.GetOrchestrator().GetContainer(DBContainerName)
	if err != nil {
		return nil, err
	}
	db, err := container.GetInstance[*postgres.PgDB](dbCnt, InstanceDB)
	if err != nil {
		return nil, err
	}
	res, err := postgres.NewTestRepository(db, db)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), fmt.Sprintf("create [%s] repo instance failed", name), err)
	}

	return repository.NewTestMetricsRepository(res), nil
}
