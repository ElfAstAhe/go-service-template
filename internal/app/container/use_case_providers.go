package container

import (
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

//goland:noinspection DuplicatedCode
func (ucc *UseCaseContainer) providerTM() (any, error) {
	dbInst, err := container.GetInstance[db.DB](InstanceDB)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return db.NewTxManager(dbInst), nil
}

//goland:noinspection DuplicatedCode
func (ucc *UseCaseContainer) providerTestGetUC() (any, error) {
	repoTest, err := container.GetInstance[domain.TestRepository](InstanceTestRepo)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return usecase.NewTestGetUseCase(repoTest), nil
}

func (ucc *UseCaseContainer) providerTestGetByCodeUC() (any, error) {
	repoTest, err := container.GetInstance[domain.TestRepository](InstanceTestRepo)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return usecase.NewTestGetByCodeUseCase(repoTest), nil
}

func (ucc *UseCaseContainer) providerTestListUC() (any, error) {
	repoTest, err := container.GetInstance[domain.TestRepository](InstanceTestRepo)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return usecase.NewTestListUseCase(repoTest), nil
}

//goland:noinspection DuplicatedCode
func (ucc *UseCaseContainer) providerTestSaveUC() (any, error) {
	trMan, err := container.GetInstance[db.TransactionManager](InstanceTM)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}
	repoTest, err := container.GetInstance[domain.TestRepository](InstanceTestRepo)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return usecase.NewTestSaveUseCase(trMan, repoTest), nil
}

//goland:noinspection DuplicatedCode
func (ucc *UseCaseContainer) providerTestDeleteUC() (any, error) {
	trMan, err := container.GetInstance[db.TransactionManager](InstanceTM)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}
	repoTest, err := container.GetInstance[domain.TestRepository](InstanceTestRepo)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return usecase.NewTestDeleteUseCase(trMan, repoTest), nil
}
