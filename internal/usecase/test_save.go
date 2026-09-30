package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type TestSaveUseCase interface {
	Save(context.Context, *domain.Test) (*domain.Test, error)
}

type TestSaveInteractor struct {
	uw   pkgdom.UnitOfWork
	repo domain.TestRepository
}

var _ TestSaveUseCase = (*TestSaveInteractor)(nil)

func NewTestSaveUseCase(uw pkgdom.UnitOfWork, repo domain.TestRepository) *TestSaveInteractor {
	return &TestSaveInteractor{
		uw:   uw,
		repo: repo,
	}
}

func (ts *TestSaveInteractor) Save(ctx context.Context, model *domain.Test) (*domain.Test, error) {
	var res *domain.Test
	err := ts.uw.Execute(ctx, func(ctx context.Context) error {
		var txErr error
		if !model.IsExists() {
			res, txErr = ts.repo.Create(ctx, model)
		} else {
			res, txErr = ts.repo.Change(ctx, model)
		}
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); ok {
			return nil, errs.NewBllNotFoundError("TestSaveUseCase.Save", "Test", model.Code, err)
		}
		if _, ok := errors.AsType[*errs.DalAlreadyExistsError](err); ok {
			return nil, errs.NewBllUniqueError("TestSaveUseCase.Save", "Test", model.Code, err)
		}

		return nil, errs.NewBllError("TestSaveUseCase.Save", fmt.Sprintf("save test model id [%v] failed", model.GetID()), err)
	}

	return res, nil
}
