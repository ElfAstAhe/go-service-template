package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	domerrs "github.com/ElfAstAhe/go-service-template/internal/domain/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type TestGetByCodeUseCase interface {
	Get(context.Context, string) (*domain.Test, error)
}

type TestGetByCodeInteractor struct {
	repo domain.TestRepository
}

func NewTestGetByCodeUseCase(repo domain.TestRepository) *TestGetByCodeInteractor {
	return &TestGetByCodeInteractor{repo: repo}
}

func (tgc *TestGetByCodeInteractor) Get(ctx context.Context, code string) (*domain.Test, error) {
	res, err := tgc.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.As(err, new(*errs.DalNotFoundError)) {
			return nil, domerrs.NewBllNotFoundError("TestGetByCodeInteractor.Get", "Test", code, err)
		}

		return nil, domerrs.NewBllError("TestGetByCodeInteractor.Get", fmt.Sprintf("find test model code [%s] failed", code), err)
	}

	return res, nil
}
