package repository

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type TestTraceRepository struct {
	*repository.BaseCRUDTraceRepository[*domain.Test, string]
	repo domain.TestRepository
}

var _ domain.TestRepository = (*TestTraceRepository)(nil)

func NewTestTraceRepository(repo domain.TestRepository) *TestTraceRepository {
	return &TestTraceRepository{
		BaseCRUDTraceRepository: repository.NewBaseCRUDTraceRepository("TestRepository", repo),
		repo:                    repo,
	}
}

func (ttr *TestTraceRepository) FindByCode(ctx context.Context, code string) (*domain.Test, error) {
	ctx, span := ttr.GetTracer().Start(ctx, fmt.Sprintf("%s.FindByCode", ttr.BaseCRUDTraceRepository.GetRepositoryName()))
	span.SetAttributes(attribute.String("param.code", code))
	defer span.End()

	res, err := ttr.repo.FindByCode(ctx, code)
	if err != nil {
		span.AddEvent("findByCode_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return res, nil
}
