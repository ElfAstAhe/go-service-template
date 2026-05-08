package container

import (
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
)

func (ucc *UseCaseContainer) providerTxManager(name string) (any, error) {
	dbCnt, err := ucc.GetOrchestrator().GetContainer(DBContainerName)
	if err != nil {
		return nil, err
	}
	dbInst, err := container.GetInstance[db.DB](dbCnt, InstanceDB)
	if err != nil {
		return nil, err
	}

	return db.NewTxManager(dbInst), nil
}

func (ucc *UseCaseContainer) providerTestGetUC(name string) (any, error) {
	repoCnt, err := ucc.GetOrchestrator().GetContainer(RepositoryContainerName)
	if err != nil {
		return nil, err
	}
	repoTest, err := container.GetInstance[domain.TestRepository](repoCnt, InstanceTestRepo)
	if err != nil {
		return nil, err
	}

	return usecase.NewTestGetUseCase(repoTest), nil
}

func (ucc *UseCaseContainer) providerTestGetByCodeUC(name string) (any, error) {
	repoCnt, err := ucc.GetOrchestrator().GetContainer(RepositoryContainerName)
	if err != nil {
		return nil, err
	}
	repoTest, err := container.GetInstance[domain.TestRepository](repoCnt, InstanceTestRepo)
	if err != nil {
		return nil, err
	}

	return usecase.NewTestGetByCodeUseCase(repoTest), nil
}

func (ucc *UseCaseContainer) providerTestListUC(name string) (any, error) {
	repoCnt, err := ucc.GetOrchestrator().GetContainer(RepositoryContainerName)
	if err != nil {
		return nil, err
	}
	repoTest, err := container.GetInstance[domain.TestRepository](repoCnt, InstanceTestRepo)
	if err != nil {
		return nil, err
	}

	return usecase.NewTestListUseCase(repoTest), nil
}

//goland:noinspection DuplicatedCode
func (ucc *UseCaseContainer) providerTestSaveUC(name string) (any, error) {
	trMan, err := container.GetInstance[db.TransactionManager](ucc, InstanceTransactionManager)
	if err != nil {
		return nil, err
	}
	repoCnt, err := ucc.GetOrchestrator().GetContainer(RepositoryContainerName)
	if err != nil {
		return nil, err
	}
	repoTest, err := container.GetInstance[domain.TestRepository](repoCnt, InstanceTestRepo)
	if err != nil {
		return nil, err
	}

	return usecase.NewTestSaveUseCase(trMan, repoTest), nil
}

//goland:noinspection DuplicatedCode
func (ucc *UseCaseContainer) providerTestDeleteUC(name string) (any, error) {
	trMan, err := container.GetInstance[db.TransactionManager](ucc, InstanceTransactionManager)
	if err != nil {
		return nil, err
	}
	repoCnt, err := ucc.GetOrchestrator().GetContainer(RepositoryContainerName)
	if err != nil {
		return nil, err
	}
	repoTest, err := container.GetInstance[domain.TestRepository](repoCnt, InstanceTestRepo)
	if err != nil {
		return nil, err
	}

	return usecase.NewTestDeleteUseCase(trMan, repoTest), nil
}
