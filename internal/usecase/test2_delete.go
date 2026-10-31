package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type Test2DeleteUseCase interface {
	Delete(context.Context, string) error
}

type Test2DeleteInteractor struct {
	uw   pkgdom.UnitOfWork
	repo domain.Test2Repository
}

var _ Test2DeleteUseCase = (*Test2DeleteInteractor)(nil)

func NewTest2Delete(uw pkgdom.UnitOfWork, repo domain.Test2Repository) *Test2DeleteInteractor {
	return &Test2DeleteInteractor{
		uw:   uw,
		repo: repo,
	}
}

func (t2d *Test2DeleteInteractor) Delete(ctx context.Context, id string) error {
	err := t2d.uw.Execute(ctx, func(uwCtx context.Context) error {
		return t2d.repo.Delete(uwCtx, id)
	})
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); ok {
			return errs.NewBllNotFoundError("Test2DeleteInteractor.Delete", "Test2", id, err)
		}

		return errs.NewBllError("Test2DeleteInteractor.Delete", fmt.Sprintf("delete test2 model id [%s] failed", id), err)
	}

	return nil
}
