package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type Test2GetByCode interface {
	Get(context.Context, string) (*domain.Test2, error)
}

type Test2GetByCodeInteractor struct {
	repo domain.Test2Repository
}

var _ Test2GetByCode = (*Test2GetByCodeInteractor)(nil)

func NewTest2GetByCode(repo domain.Test2Repository) *Test2GetByCodeInteractor {
	return &Test2GetByCodeInteractor{
		repo: repo,
	}
}

func (t2c *Test2GetByCodeInteractor) Get(ctx context.Context, code string) (*domain.Test2, error) {
	res, err := t2c.repo.FindByCode(ctx, code)
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); ok {
			return nil, errs.NewBllNotFoundError("Test2GetByCodeInteractor.Get", "Test2", code, err)
		}

		return nil, errs.NewBllError("Test2GetByCodeInteractor.Get", fmt.Sprintf("find test2 model code [%s] failed", code), err)
	}

	return res, nil
}
