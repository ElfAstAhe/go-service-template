package usecase

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
)

type TestCreateUseCase struct {
	repo domain.TestRepository
}

func NewTestCreateUseCase(repo domain.TestRepository) *TestCreateUseCase {
	return &TestCreateUseCase{
		repo: repo,
	}
}

func (tc *TestCreateUseCase) Create(ctx context.Context, test *domain.Test) (*domain.Test, error) {
	return tc.repo.Create(ctx, test)
}
