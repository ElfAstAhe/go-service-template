package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type Test2SaveUseCase interface {
	Save(context.Context, *domain.Test2) (*domain.Test2, error)
}

type Test2SaveInteractor struct {
	uw   pkgdom.UnitOfWork
	repo domain.Test2Repository
}

var _ Test2SaveUseCase = (*Test2SaveInteractor)(nil)

func NewTest2Save(uw pkgdom.UnitOfWork, repo domain.Test2Repository) *Test2SaveInteractor {
	return &Test2SaveInteractor{
		uw:   uw,
		repo: repo,
	}
}

func (t2s *Test2SaveInteractor) Save(ctx context.Context, model *domain.Test2) (*domain.Test2, error) {
	var res *domain.Test2
	err := t2s.uw.Execute(ctx, func(uwCtx context.Context) error {
		var uwErr error
		if !model.IsExists() {
			res, uwErr = t2s.repo.Create(uwCtx, model)
		} else {
			res, uwErr = t2s.repo.Change(uwCtx, model)
		}

		return uwErr
	})
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); ok {
			return nil, errs.NewBllNotFoundError("Test2SaveInteractor.Save", "Test2", model.Code, err)
		}
		if _, ok := errors.AsType[*errs.DalAlreadyExistsError](err); ok {
			return nil, errs.NewBllUniqueError("TestSaveInteractor.Save", "Test2", model.Code, err)
		}

		return nil, errs.NewBllError("TestSaveInteractor.Save", fmt.Sprintf("save test2 model id [%v] failed", model.GetID()), err)
	}

	return res, nil
}
