package usecase

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/domain/errs"
)

type TestGetUseCase interface {
	Get(ctx context.Context, id string) (*domain.Test, error)
}

type TestGetUseCaseImpl struct {
	repo domain.TestRepository
}

func NewTestGetUseCase(repo domain.TestRepository) *TestGetUseCaseImpl {
	return &TestGetUseCaseImpl{
		repo: repo,
	}
}

func (tg *TestGetUseCaseImpl) Get(ctx context.Context, id string) (*domain.Test, error) {
	res, err := tg.repo.Find(ctx, id)
	if err != nil {
		return nil, errs.NewBllError("TestGetUseCaseImpl", fmt.Sprintf("find test model id [%s] failed", id), err)
	}

	return res, nil
}
