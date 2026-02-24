package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
)

const (
	sqlTestFind = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test
where
    id = $1
`
	sqlTestFindByCode = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test
where
    code = $1
`
	sqlTestCreate = `
insert into test (
    id,
    code,
    name,
    description,
    created_at,
    modified_at
)
values (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6
)
returning
    id,
    code,
    name,
    description,
    created_at,
    modified_at
`
	sqlTestChange = `
update
    test
set
    code = $2,
    name = $3,
    description = $4,
    modified_at = $5
where
    id = $1
`
	sqlTestDelete = `
delete
from
    test
where
    id = $1
`
)

type TestRepositoryImpl struct {
	*repository.BaseRepository[*domain.Test, string]
	onceFindByCode sync.Once
	findByCodeStmt *sql.Stmt
}

func NewTestRepository(db db.DB) (*TestRepositoryImpl, error) {
	res := &TestRepositoryImpl{}

	queryBuilders := repository.NewBaseQueryBuildersBuilder().NewInstance().
		WithFind(func() string {
			return sqlTestFind
		}).
		WithCreate(func() string {
			return sqlTestCreate
		}).
		WithChange(func() string {
			return sqlTestChange
		}).
		WithDelete(func() string {
			return sqlTestDelete
		}).
		WithList(func() string {
			return ""
		}).
		Build()

	callbacks, err := repository.NewBaseRepositoryCallbacksBuilder[*domain.Test, string]().NewInstance().
		WithNewEntityFactory()
	WithAfterFind(nil).
		Build()
	if err != nil {
		return nil, errs.NewCommonError("error create test repo callbacks", err)
	}

	base, err := repository.NewBaseRepository[*domain.Test, string](
		db,
		repository.NewEntityInfo("test", "Test"),
		queryBuilders,
		callbacks,
	)
	if err != nil {
		return nil, errs.NewCommonError("error create TestRepository", err)
	}

	res.BaseRepository = base

	return res, nil
}

func (tr *TestRepositoryImpl) FindByCode(ctx context.Context, code string) (*domain.Test, error) {
	if err := tr.prepareFindByCode(); err != nil {
		return nil, errs.NewNotImplementedError(err)
	}

	res, err := tr.GetCallbacks().RowScanner(tr.findByCodeStmt.QueryRowContext(ctx, code))
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return nil, errs.NewDalNotFoundError(tr.GetInfo().Entity, code, err)
		}

		return nil, errs.NewDalError("BaseRepository.Find", "get row", err)
	}

	if tr.GetCallbacks().AfterFind != nil {
		return tr.GetCallbacks().AfterFind(res)
	}

	return res, nil
}

func (tr *TestRepositoryImpl) prepareFindByCode() error {
	var err error = nil
	tr.onceFindByCode.Do(func() {
		queryCtx, queryCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer queryCancel()

		tr.findByCodeStmt, err = tr.GetDB().GetDB().PrepareContext(queryCtx, sqlTestFindByCode)
		if err != nil {
			err = errs.NewDalError("TestRepository.prepareFindByCode", "prepare find by code stmt", err)
		}
	})
	if err != nil {
		return errs.NewDalError("TestRepository.prepareFindByCode", "prepare find by code", err)
	}

	return nil
}
