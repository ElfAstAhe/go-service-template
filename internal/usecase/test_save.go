package usecase

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/domain/errs"
	usecase "github.com/ElfAstAhe/go-service-template/pkg/db"
)

type TestSaveUseCase interface {
	Save(context.Context, *domain.Test) (*domain.Test, error)
}

type TestSaveInteractor struct {
	tm   usecase.TransactionManager
	repo domain.TestRepository
}

func NewTestSaveUseCase(tm usecase.TransactionManager, repo domain.TestRepository) *TestSaveInteractor {
	return &TestSaveInteractor{
		tm:   tm,
		repo: repo,
	}
}

func (ts *TestSaveInteractor) Save(ctx context.Context, model *domain.Test) (*domain.Test, error) {
	var res *domain.Test
	err := ts.tm.WithinTransaction(ctx, nil, func(ctx context.Context) error {
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
		return nil, errs.NewBllError("TestSaveUseCase.Save", fmt.Sprintf("save test model id [%v] failed", model.GetID()), err)
	}

	return res, nil
}
