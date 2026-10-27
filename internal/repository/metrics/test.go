package metrics

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository/metrics"
)

type TestRepository struct {
	*metrics.CRUD[*domain.Test, string]
	repo domain.TestRepository
}

var _ pkgdom.CRUDRepository[*domain.Test, string] = (*TestRepository)(nil)
var _ domain.TestRepository = (*TestRepository)(nil)

func NewTestRepository(repo domain.TestRepository) *TestRepository {
	return &TestRepository{
		CRUD: metrics.NewCRUD("TestRepository", repo),
		repo: repo,
	}
}

func (tmr *TestRepository) FindByCode(ctx context.Context, code string) (res *domain.Test, err error) {
	defer func(start time.Time) {
		metrics.ObserveRepositoryOp(tmr.GetRepositoryName(), "FindByCode", err, start)
	}(time.Now())

	return tmr.repo.FindByCode(ctx, code)
}
