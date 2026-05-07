package container

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/repository"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
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
	return nil
}

func (rc *RepositoryContainer) providerTestRepository(name string) (any, error) {
	dbCnt, err := rc.GetOrchestrator().GetContainer(DBContainerName)
	if err != nil {
		return nil, err
	}
	database, err := container.GetInstance[*postgres.PgDB](dbCnt, InstanceDB)
	if err != nil {
		return nil, err
	}
	res, err := postgres.NewTestRepository(database, database)
	if err != nil {
		return nil, errs.NewContainerError(rc.GetName(), fmt.Sprintf("create [%s] repo instance failed", name), err)
	}

	return repository.NewTestMetricsRepository(res), nil
}
