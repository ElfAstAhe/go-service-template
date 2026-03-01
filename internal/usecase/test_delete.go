package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	domerrs "github.com/ElfAstAhe/go-service-template/internal/domain/errs"
	usecase "github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type TestDeleteUseCase interface {
	Delete(context.Context, string) error
}

type TestDeleteInteractor struct {
	tm   usecase.TransactionManager
	repo domain.TestRepository
}

func NewTestDeleteUseCase(tm usecase.TransactionManager, repo domain.TestRepository) *TestDeleteInteractor {
	return &TestDeleteInteractor{
		tm:   tm,
		repo: repo,
	}
}

func (td *TestDeleteInteractor) Delete(ctx context.Context, id string) error {
	err := td.tm.WithinTransaction(ctx, nil, func(ctx context.Context) error {
		return td.repo.Delete(ctx, id)
	})
	if err != nil {
		if errors.As(err, new(*errs.DalNotFoundError)) {
			return domerrs.NewBllNotFoundError("TestDeleteInteractor.Delete", "Test", id, err)
		}

		return domerrs.NewBllError("TestDeleteInteractor.Delete", fmt.Sprintf("delete test model id [%s] failed", id), err)
	}

	return nil
}
