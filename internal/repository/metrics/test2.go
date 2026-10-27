package metrics

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository/metrics"
)

type Test2Repository struct {
	*metrics.CRUD[*domain.Test2, string]
	next domain.Test2Repository
}

var _ pkgdom.CRUDRepository[*domain.Test2, string] = (*Test2Repository)(nil)
var _ domain.Test2Repository = (*Test2Repository)(nil)

func NewTest2Repository(repo domain.Test2Repository) *Test2Repository {
	return &Test2Repository{
		CRUD: metrics.NewCRUD("Test2", repo),
		next: repo,
	}
}

func (t2r *Test2Repository) FindByCode(ctx context.Context, code string) (res *domain.Test2, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(t2r.GetRepositoryName(), "FindByCode", err, start)
	}(time.Now())

	return t2r.next.FindByCode(ctx, code)
}
