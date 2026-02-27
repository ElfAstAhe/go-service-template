package usecase

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/domain/errs"
	usecase "github.com/ElfAstAhe/go-service-template/pkg/db"
)

type TestDeleteUseCase interface {
	Delete(context.Context, string) error
}

type TestDeleteUseCaseImpl struct {
	tm   usecase.TransactionManager
	repo domain.TestRepository
}

func NewTestDeleteUseCase(tm usecase.TransactionManager, repo domain.TestRepository) *TestDeleteUseCaseImpl {
	return &TestDeleteUseCaseImpl{
		tm:   tm,
		repo: repo,
	}
}

func (td *TestDeleteUseCaseImpl) Delete(ctx context.Context, id string) error {
	err := td.tm.WithinTransaction(ctx, nil, func(ctx context.Context) error {
		if err := td.repo.Delete(ctx, id); err != nil {
			return errs.NewBllError("TestDeleteUseCaseImpl.Delete", "run in transaction", err)
		}

		return nil
	})
	if err != nil {
		return errs.NewBllError("TestDeleteUseCaseImpl.Delete", fmt.Sprintf("delete test model id [%s] failed", id), err)
	}

	return nil
}
