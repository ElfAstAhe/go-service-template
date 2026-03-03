package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	domerrs "github.com/ElfAstAhe/go-service-template/internal/domain/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type TestGetUseCase interface {
	Get(ctx context.Context, id string) (*domain.Test, error)
}

type TestGetInteractor struct {
	repo domain.TestRepository
}

func NewTestGetUseCase(repo domain.TestRepository) *TestGetInteractor {
	return &TestGetInteractor{
		repo: repo,
	}
}

func (tg *TestGetInteractor) Get(ctx context.Context, id string) (*domain.Test, error) {
	res, err := tg.repo.Find(ctx, id)
	if err != nil {
		if errors.As(err, new(*errs.DalNotFoundError)) {
			return nil, domerrs.NewBllNotFoundError("TestGetInteractor.Get", "Test", id, err)
		}

		return nil, domerrs.NewBllError("TestGetInteractor.Get", fmt.Sprintf("find test model id [%s] failed", id), err)
	}

	return res, nil
}
