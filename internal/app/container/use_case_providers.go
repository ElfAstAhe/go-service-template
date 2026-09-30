package container

import (
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkguc "github.com/ElfAstAhe/go-service-template/pkg/usecase"
)

func (ucc *UseCaseContainer) providerUnitOfWork() (any, error) {
	tmInst, err := container.GetInstance[db.TransactionManager](InstanceTM)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}
	tmOptsInst, err := container.GetInstance[*db.TransactionOptions](InstanceTMOpts)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return pkguc.NewUnitOfWork(tmInst, tmOptsInst), nil
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
	uwInst, err := container.GetInstance[pkgdom.UnitOfWork](InstanceUnitOfWork)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}
	repoTest, err := container.GetInstance[domain.TestRepository](InstanceTestRepo)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return usecase.NewTestSaveUseCase(uwInst, repoTest), nil
}

//goland:noinspection DuplicatedCode
func (ucc *UseCaseContainer) providerTestDeleteUC() (any, error) {
	uwInst, err := container.GetInstance[pkgdom.UnitOfWork](InstanceUnitOfWork)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}
	repoTest, err := container.GetInstance[domain.TestRepository](InstanceTestRepo)
	if err != nil {
		return nil, errs.NewContainerError(ucc.GetName(), "provider: retrieve instance failed", err)
	}

	return usecase.NewTestDeleteUseCase(uwInst, repoTest), nil
}
