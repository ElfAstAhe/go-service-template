package trace

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Test2Repository struct {
	*trace.CRUD[*domain.Test2, string]
	next domain.Test2Repository
}

var _ pkgdom.CRUDRepository[*domain.Test2, string] = (*Test2Repository)(nil)
var _ domain.Test2Repository = (*Test2Repository)(nil)

func NewTest2Repository(next domain.Test2Repository) (*Test2Repository, error) {
	return &Test2Repository{
		CRUD: trace.NewCRUD("Test2", next),
		next: next,
	}, nil
}

//goland:noinspection DuplicatedCode
func (t2t Test2Repository) FindByCode(ctx context.Context, code string) (*domain.Test2, error) {
	ctx, span := t2t.GetTracer().Start(ctx, fmt.Sprintf("%s.FindByCode", t2t.GetRepositoryName()))
	span.SetAttributes(attribute.String("param.code", code))
	defer span.End()

	res, err := t2t.next.FindByCode(ctx, code)
	if err != nil {
		span.AddEvent("findByCode_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return res, nil
}
