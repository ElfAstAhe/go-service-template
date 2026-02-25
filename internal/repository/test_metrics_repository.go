package repository

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/metrics"
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
)

type TestMetricsRepository struct {
	*repository.BaseMetricsRepository[*domain.Test, string]
	repo domain.TestRepository
}

func NewTestMetricsRepository(repo domain.TestRepository) *TestMetricsRepository {
	return &TestMetricsRepository{
		BaseMetricsRepository: repository.NewBaseMetricsRepository(repo),
		repo:                  repo,
	}
}

func (tmr *TestMetricsRepository) FindByCode(ctx context.Context, code string) (res *domain.Test, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(tmr.BaseMetricsRepository.GetRepositoryName(), "FindByCode", err, start)
	}(time.Now())

	return tmr.repo.FindByCode(ctx, code)
}
