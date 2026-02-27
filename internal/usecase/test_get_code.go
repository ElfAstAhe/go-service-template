package usecase

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/domain/errs"
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
		return nil, errs.NewBllError("TestGetByCodeInteractor.Get", fmt.Sprintf("find test model code [%s] failed", code), err)
	}

	return res, nil
}
