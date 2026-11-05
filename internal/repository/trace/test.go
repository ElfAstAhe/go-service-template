package trace

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	domain2 "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type TestRepository struct {
	*trace.CRUD[*domain.Test, string]
	repo domain.TestRepository
}

var _ domain2.CRUDRepository[*domain.Test, string] = (*TestRepository)(nil)
var _ domain.TestRepository = (*TestRepository)(nil)

func NewTestRepository(repo domain.TestRepository) *TestRepository {
	return &TestRepository{
		CRUD: trace.NewCRUD("TestRepository", repo),
		repo: repo,
	}
}

//goland:noinspection DuplicatedCode
func (ttr *TestRepository) FindByCode(ctx context.Context, code string) (*domain.Test, error) {
	ctx, span := ttr.GetTracer().Start(ctx, fmt.Sprintf("%s.FindByCode", ttr.GetRepositoryName()))
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
