package repository

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type BaseCRUDTraceRepository[T domain.Entity[ID], ID any] struct {
	repository domain.CrudRepository[T, ID]
	repoName   string
	tracer     trace.Tracer
	nilEntity  T
}

func NewBaseCRUDTraceRepository[T domain.Entity[ID], ID any](repositoryName string, repository domain.CrudRepository[T, ID]) *BaseCRUDTraceRepository[T, ID] {
	return &BaseCRUDTraceRepository[T, ID]{
		repository: repository,
		repoName:   repositoryName,
		tracer:     otel.GetTracerProvider().Tracer(repositoryName),
	}
}

func (btr *BaseCRUDTraceRepository[T, ID]) Find(ctx context.Context, id ID) (T, error) {
	ctx, span := btr.tracer.Start(ctx, fmt.Sprintf("%s.Find", btr.repoName))
	span.SetAttributes(attribute.String("param.id", fmt.Sprintf("%v", id)))
	defer span.End()

	res, err := btr.repository.Find(ctx, id)
	if err != nil {
		span.AddEvent("find_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return btr.nilEntity, err
	}

	return res, nil
}

func (btr *BaseCRUDTraceRepository[T, ID]) List(ctx context.Context, limit, offset int) ([]T, error) {
	ctx, span := btr.tracer.Start(ctx, fmt.Sprintf("%s.List", btr.repoName))
	span.SetAttributes(attribute.Int("param.limit", limit))
	span.SetAttributes(attribute.Int("param.offset", offset))
	defer span.End()

	res, err := btr.repository.List(ctx, limit, offset)
	if err != nil {
		span.AddEvent("list_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return res, nil
}

func (btr *BaseCRUDTraceRepository[T, ID]) Create(ctx context.Context, entity T) (T, error) {
	ctx, span := btr.tracer.Start(ctx, fmt.Sprintf("%s.Create", btr.repoName))
	span.SetAttributes(attribute.String("param.entity_id", fmt.Sprintf("%v", entity.GetID())))
	defer span.End()

	res, err := btr.repository.Create(ctx, entity)
	if err != nil {
		span.AddEvent("create_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return btr.nilEntity, err
	}

	return res, nil
}

func (btr *BaseCRUDTraceRepository[T, ID]) Change(ctx context.Context, entity T) (T, error) {
	ctx, span := btr.tracer.Start(ctx, fmt.Sprintf("%s.Change", btr.repoName))
	span.SetAttributes(attribute.String("param.entity_id", fmt.Sprintf("%v", entity.GetID())))
	defer span.End()

	res, err := btr.repository.Change(ctx, entity)
	if err != nil {
		span.AddEvent("change_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return btr.nilEntity, err
	}

	return res, nil
}

func (btr *BaseCRUDTraceRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	ctx, span := btr.tracer.Start(ctx, fmt.Sprintf("%s.Delete", btr.repoName))
	span.SetAttributes(attribute.String("param.id", fmt.Sprintf("%v", id)))
	defer span.End()

	err := btr.repository.Delete(ctx, id)
	if err != nil {
		span.AddEvent("delete_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

func (btr *BaseCRUDTraceRepository[T, ID]) GetRepositoryName() string {
	return btr.repoName
}

func (btr *BaseCRUDTraceRepository[T, ID]) GetTracer() trace.Tracer {
	return btr.tracer
}
