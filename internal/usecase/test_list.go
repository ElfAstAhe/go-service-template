package usecase

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/domain/errs"
)

type TestListUseCase interface {
	List(ctx context.Context, limit, offset int) ([]*domain.Test, error)
}

type TestListUseCaseImpl struct {
	repo domain.TestRepository
}

func NewTestListUseCase(repo domain.TestRepository) *TestListUseCaseImpl {
	return &TestListUseCaseImpl{
		repo: repo,
	}
}

func (tl *TestListUseCaseImpl) List(ctx context.Context, limit, offset int) ([]*domain.Test, error) {
	res, err := tl.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, errs.NewBllError("TestListUseCase.List", fmt.Sprintf("list test data with limit [%v] and offset [%v] failed", limit, offset), err)
	}

	return res, nil
}
