package repository

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/metrics"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type BaseCRUDMetricsRepository[T domain.Entity[ID], ID comparable] struct {
	repository domain.CRUDRepository[T, ID]
	repoName   string
}

func NewBaseCRUDMetricsRepository[T domain.Entity[ID], ID comparable](repoName string, repository domain.CRUDRepository[T, ID]) *BaseCRUDMetricsRepository[T, ID] {
	res := &BaseCRUDMetricsRepository[T, ID]{
		repository: repository,
		repoName:   repoName,
	}
	if repoName == "" {
		res.repoName = utils.GetTypeName(repository)
	}

	return res
}

func (bmr *BaseCRUDMetricsRepository[T, ID]) Find(ctx context.Context, id ID) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Find", err, start)
	}(time.Now())

	return bmr.repository.Find(ctx, id)
}

func (bmr *BaseCRUDMetricsRepository[T, ID]) List(ctx context.Context, limit, offset int) (res []T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "List", err, start)
	}(time.Now())

	return bmr.repository.List(ctx, limit, offset)
}

func (bmr *BaseCRUDMetricsRepository[T, ID]) Create(ctx context.Context, entity T) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Create", err, start)
	}(time.Now())

	return bmr.repository.Create(ctx, entity)
}

func (bmr *BaseCRUDMetricsRepository[T, ID]) Change(ctx context.Context, entity T) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Change", err, start)
	}(time.Now())

	return bmr.repository.Change(ctx, entity)
}

func (bmr *BaseCRUDMetricsRepository[T, ID]) Delete(ctx context.Context, id ID) (err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(bmr.repoName, "Delete", err, start)
	}(time.Now())

	return bmr.repository.Delete(ctx, id)
}

func (bmr *BaseCRUDMetricsRepository[T, ID]) GetRepositoryName() string {
	return bmr.repoName
}
