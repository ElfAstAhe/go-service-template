package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type Test2GetUseCase interface {
	Get(context.Context, string) (*domain.Test2, error)
}

type Test2GetInteractor struct {
	repo domain.Test2Repository
}

var _ Test2GetUseCase = (*Test2GetInteractor)(nil)

func NewTest2Get(repo domain.Test2Repository) *Test2GetInteractor {
	return &Test2GetInteractor{
		repo: repo,
	}
}

func (t2g *Test2GetInteractor) Get(ctx context.Context, id string) (*domain.Test2, error) {
	res, err := t2g.repo.Find(ctx, id)
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); ok {
			return nil, errs.NewBllNotFoundError("Test2GetInteractor.Get", "Test2", id, err)
		}

		return nil, errs.NewBllError("Test2GetInteractor.Get", fmt.Sprintf("find test2 model id [%s] failed", id), err)
	}

	return res, nil
}
