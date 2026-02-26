package usecase

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	usecase "github.com/ElfAstAhe/go-service-template/pkg/db"
)

type TestChangeUseCase struct {
	tm   usecase.TransactionManager
	repo domain.TestRepository
}

func NewTestChangeUseCase(tm usecase.TransactionManager, repo domain.TestRepository) *TestChangeUseCase {
	return &TestChangeUseCase{
		tm:   tm,
		repo: repo,
	}
}

func (tcu *TestChangeUseCase) Change(ctx context.Context, test *domain.Test) (*domain.Test, error) {
	return tcu.repo.Change(ctx, test)
}
