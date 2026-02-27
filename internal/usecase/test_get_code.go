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

type TestGetByCodeUseCaseImpl struct {
	repo domain.TestRepository
}

func NewTestGetCodeUseCase(repo domain.TestRepository) *TestGetByCodeUseCaseImpl {
	return &TestGetByCodeUseCaseImpl{repo: repo}
}

func (tgc *TestGetByCodeUseCaseImpl) Get(ctx context.Context, code string) (*domain.Test, error) {
	res, err := tgc.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, errs.NewBllError("TestGetByCodeUseCaseImpl.Get", fmt.Sprintf("find test model code [%s] failed", code), err)
	}

	return res, nil
}
