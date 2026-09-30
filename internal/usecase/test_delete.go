package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type TestDeleteUseCase interface {
	Delete(context.Context, string) error
}

type TestDeleteInteractor struct {
	uw   pkgdom.UnitOfWork
	repo domain.TestRepository
}

var _ TestDeleteUseCase = (*TestDeleteInteractor)(nil)

func NewTestDeleteUseCase(
	uw pkgdom.UnitOfWork,
	repo domain.TestRepository,
) *TestDeleteInteractor {
	return &TestDeleteInteractor{
		uw:   uw,
		repo: repo,
	}
}

func (td *TestDeleteInteractor) Delete(ctx context.Context, id string) error {
	err := td.uw.Execute(ctx, func(ctx context.Context) error {
		return td.repo.Delete(ctx, id)
	})
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); ok {
			return errs.NewBllNotFoundError("TestDeleteInteractor.Delete", "Test", id, err)
		}

		return errs.NewBllError("TestDeleteInteractor.Delete", fmt.Sprintf("delete test model id [%s] failed", id), err)
	}

	return nil
}
