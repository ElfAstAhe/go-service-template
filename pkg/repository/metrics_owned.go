package repository

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/metrics"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type BaseOwnedMetricsRepository[T domain.Entity[ID], ID comparable, OwnerID comparable] struct {
	repository domain.OwnedRepository[T, ID, OwnerID]
	repoName   string
}

func NewBaseOwnedMetricsRepository[T domain.Entity[ID], ID comparable, OwnerID comparable](repoName string, repository domain.OwnedRepository[T, ID, OwnerID]) *BaseOwnedMetricsRepository[T, ID, OwnerID] {
	res := &BaseOwnedMetricsRepository[T, ID, OwnerID]{
		repoName:   repoName,
		repository: repository,
	}
	if repoName != "" {
		res.repoName = utils.GetTypeName(repository)
	}

	return res
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Find(ctx context.Context, ownerID OwnerID, id ID) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "Find", err, start)
	}(time.Now())

	return omr.repository.Find(ctx, ownerID, id)
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) List(ctx context.Context, ownerID OwnerID, limit, offset int) (res []T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "List", err, start)
	}(time.Now())

	return omr.repository.List(ctx, ownerID, limit, offset)
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) ListAll(ctx context.Context, ownerID OwnerID) (res []T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "ListAll", err, start)
	}(time.Now())

	return omr.repository.ListAll(ctx, ownerID)
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) ListAllByOwners(ctx context.Context, ownerIDs OwnerID) (res map[OwnerID][]T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "ListAllByOwners", err, start)
	}(time.Now())

	return omr.repository.ListAllByOwners(ctx, ownerIDs)
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Save(ctx context.Context, ownerID OwnerID, owned []T) (res []T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "Save", err, start)
	}(time.Now())

	return omr.repository.Save(ctx, ownerID, owned)
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Create(ctx context.Context, ownerID OwnerID, entity T) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "Create", err, start)
	}(time.Now())

	return omr.repository.Create(ctx, ownerID, entity)
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Change(ctx context.Context, ownerID OwnerID, entity T) (res T, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "Change", err, start)
	}(time.Now())

	return omr.repository.Change(ctx, ownerID, entity)
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) DeleteAll(ctx context.Context, ownerID OwnerID) (err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "DeleteAll", err, start)
	}(time.Now())

	return omr.repository.DeleteAll(ctx, ownerID)
}

func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Delete(ctx context.Context, ownerID OwnerID, id ID) (err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(omr.repoName, "Delete", err, start)
	}(time.Now())

	return omr.repository.Delete(ctx, ownerID, id)
}
