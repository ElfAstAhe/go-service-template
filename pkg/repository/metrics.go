package repository

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/metrics"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type BaseMetricsRepository[T domain.Entity[ID], ID any] struct {
	repository domain.Repository[T, ID]
	repoName   string
}

func NewBaseMetricsRepository[T domain.Entity[ID], ID any](repository domain.Repository[T, ID]) *BaseMetricsRepository[T, ID] {
	return &BaseMetricsRepository[T, ID]{
		repository: repository,
		repoName:   utils.GetTypeName(repository),
	}
}

func (bmr *BaseMetricsRepository[T, ID]) Find(ctx context.Context, id ID) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Find", err, start)
	}(time.Now())

	return bmr.repository.Find(ctx, id)
}

func (bmr *BaseMetricsRepository[T, ID]) List(ctx context.Context, limit, offset int) (res []T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "List", err, start)
	}(time.Now())

	return bmr.repository.List(ctx, limit, offset)
}

func (bmr *BaseMetricsRepository[T, ID]) Create(ctx context.Context, entity T) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Create", err, start)
	}(time.Now())

	return bmr.repository.Create(ctx, entity)
}

func (bmr *BaseMetricsRepository[T, ID]) Change(ctx context.Context, entity T) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Change", err, start)
	}(time.Now())

	return bmr.repository.Change(ctx, entity)
}

func (bmr *BaseMetricsRepository[T, ID]) Delete(ctx context.Context, id ID) (err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Delete", err, start)
	}(time.Now())

	return bmr.repository.Delete(ctx, id)
}

func (bmr *BaseMetricsRepository[T, ID]) Close() (err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Close", err, start)
	}(time.Now())

	return bmr.repository.Close()
}

func (bmr *BaseMetricsRepository[T, ID]) GetRepositoryName() string {
	return bmr.repoName
}
