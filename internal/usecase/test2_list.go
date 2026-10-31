package usecase

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type Test2ListUseCase interface {
	List(context.Context, int, int) ([]*domain.Test2, error)
}

type Test2ListInteractor struct {
	repo domain.Test2Repository
}

var _ Test2ListUseCase = (*Test2ListInteractor)(nil)

func NewTest2List(repo domain.Test2Repository) *Test2ListInteractor {
	return &Test2ListInteractor{
		repo: repo,
	}
}

func (t2l *Test2ListInteractor) List(ctx context.Context, limit int, offset int) ([]*domain.Test2, error) {
	res, err := t2l.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, errs.NewBllError("Test2ListInteractor.List", fmt.Sprintf("list test data with limit [%v] and offset [%v] failed", limit, offset), err)
	}

	return res, nil
}
