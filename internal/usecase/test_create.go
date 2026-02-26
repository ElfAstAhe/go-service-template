package usecase

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/domain/errs"
	usecase "github.com/ElfAstAhe/go-service-template/pkg/db"
)

type TestCreateUseCase struct {
	tm   usecase.TransactionManager
	repo domain.TestRepository
}

func NewTestCreateUseCase(tm usecase.TransactionManager, repo domain.TestRepository) *TestCreateUseCase {
	return &TestCreateUseCase{
		tm:   tm,
		repo: repo,
	}
}

func (tc *TestCreateUseCase) Create(ctx context.Context, test *domain.Test) (*domain.Test, error) {
	var res *domain.Test
	err := tc.tm.WithinTransaction(ctx, nil, func(ctx context.Context) error {
		var txErr error
		res, txErr = tc.repo.Create(ctx, test)
		if txErr != nil {
			return errs.NewBllError("TestCreateUseCase.Create", "run in transaction", txErr)
		}

		return nil
	})
	if err != nil {
		return nil, errs.NewBllError("TestCreateUseCase.Create", "error create test", err)
	}

	return res, nil
}
