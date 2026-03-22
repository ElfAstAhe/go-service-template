package repository

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/telemetry"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type BaseOwnedTraceRepository[T domain.Entity[ID], ID comparable, OwnerID comparable] struct {
	*telemetry.BaseTelemetry
	repository domain.OwnedRepository[T, ID, OwnerID]
	nilEntity  T
}

func NewBaseOwnedTraceRepository[T domain.Entity[ID], ID comparable, OwnerID comparable](repoName string, repository domain.OwnedRepository[T, ID, OwnerID]) *BaseOwnedTraceRepository[T, ID, OwnerID] {
	res := &BaseOwnedTraceRepository[T, ID, OwnerID]{
		repository: repository,
	}
	tracerName := repoName
	if tracerName == "" {
		tracerName = utils.GetTypeName(repository)
	}
	res.BaseTelemetry = telemetry.NewBaseTelemetry(tracerName)

	return res
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) Find(ctx context.Context, ownerID OwnerID, id ID) (T, error) {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.Find", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("param.owner_id", fmt.Sprintf("%v", ownerID)),
		attribute.String("param.id", fmt.Sprintf("%v", id)),
	)

	res, err := otr.repository.Find(ctx, ownerID, id)
	if err != nil {
		span.AddEvent("Find_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return otr.GetNilEntity(), err
	}

	return res, nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) List(ctx context.Context, ownerID OwnerID, limit, offset int) ([]T, error) {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.List", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("param.owner_id", fmt.Sprintf("%v", ownerID)),
		attribute.Int("param.limit", limit),
		attribute.Int("param.offset", offset),
	)

	res, err := otr.repository.List(ctx, ownerID, limit, offset)
	if err != nil {
		span.AddEvent("List_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return res, nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) ListAll(ctx context.Context, ownerID OwnerID) ([]T, error) {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.ListAll", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(attribute.String("param.owner_id", fmt.Sprintf("%v", ownerID)))

	res, err := otr.repository.ListAll(ctx, ownerID)
	if err != nil {
		span.AddEvent("ListAll_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return res, nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) ListAllByOwners(ctx context.Context, ownerIDs ...OwnerID) (map[OwnerID][]T, error) {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.ListAllByOwners", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(attribute.String("param.owner_ids", fmt.Sprintf("%v", ownerIDs)))

	res, err := otr.repository.ListAllByOwners(ctx, ownerIDs...)
	if err != nil {
		span.AddEvent("ListAllByOwners_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return res, nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) Save(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error) {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.Save", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("param.owner_id", fmt.Sprintf("%v", ownerID)),
		attribute.Int("param.owner_cnt", len(owned)),
	)

	res, err := otr.repository.Save(ctx, ownerID, owned)
	if err != nil {
		span.AddEvent("Save_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return res, nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) Create(ctx context.Context, ownerID OwnerID, entity T) (T, error) {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.Create", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("param.owner_id", fmt.Sprintf("%v", ownerID)),
		attribute.String("param.entity_id", fmt.Sprintf("%v", entity.GetID())),
	)

	res, err := otr.repository.Create(ctx, ownerID, entity)
	if err != nil {
		span.AddEvent("Create_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return otr.GetNilEntity(), err
	}

	return res, nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) Change(ctx context.Context, ownerID OwnerID, entity T) (T, error) {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.Change", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("param.owner_id", fmt.Sprintf("%v", ownerID)),
		attribute.String("param.entity_id", fmt.Sprintf("%v", entity.GetID())),
	)

	res, err := otr.repository.Change(ctx, ownerID, entity)
	if err != nil {
		span.AddEvent("Change_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return otr.GetNilEntity(), err
	}

	return res, nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) DeleteAll(ctx context.Context, ownerID OwnerID) error {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.DeleteAll", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("param.owner_id", fmt.Sprintf("%v", ownerID)),
	)

	err := otr.repository.DeleteAll(ctx, ownerID)
	if err != nil {
		span.AddEvent("DeleteAll_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) Delete(ctx context.Context, ownerID OwnerID, id ID) error {
	ctx, span := otr.StartSpan(ctx, fmt.Sprintf("%s.Delete", otr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("param.owner_id", fmt.Sprintf("%v", ownerID)),
		attribute.String("param.id", fmt.Sprintf("%v", id)),
	)

	err := otr.repository.Delete(ctx, ownerID, id)
	if err != nil {
		span.AddEvent("Delete_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) GetRepositoryName() string {
	return otr.GetTracerName()
}

func (otr *BaseOwnedTraceRepository[T, ID, OwnerID]) GetNilEntity() T {
	return otr.nilEntity
}
