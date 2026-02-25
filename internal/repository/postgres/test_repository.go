package postgres

import (
	"context"
	"database/sql"
	"errors"
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
	sqlTestList string = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test
order by
    id asc
offset $2
limit $1
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
returning
    id,
    code,
    name,
    description,
    created_at,
    modified_at
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
	mu             sync.Mutex
	findByCodeStmt *sql.Stmt
}

func NewTestRepository(db db.DB) (*TestRepositoryImpl, error) {
	// new instance
	res := &TestRepositoryImpl{}
	// sql builders
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
			return sqlTestList
		}).
		Build()
	// callbacks
	callbacks, err := repository.NewBaseRepositoryCallbacksBuilder[*domain.Test, string]().NewInstance().
		WithEntityScanner(res.entityScanner).
		WithNewEntityFactory(domain.NewEmptyTest).
		WithValidateCreate(res.validateCreate).
		WithBeforeCreate(res.beforeCreate).
		WithCreator(res.creator).
		WithValidateChange(res.validateChange).
		WithBeforeChange(res.beforeChange).
		WithChanger(res.changer).
		Build()
	if err != nil {
		return nil, errs.NewCommonError("error create test repo callbacks", err)
	}
	// base crud
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

	row := tr.findByCodeStmt.QueryRowContext(ctx, code)

	res := tr.GetCallbacks().NewEntityFactory()

	err := tr.GetCallbacks().EntityScanner(row, res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewDalNotFoundError(tr.GetInfo().Entity, code, err)
		}

		return nil, errs.NewDalError("TestRepositoryImpl.FindByCode", "get row", err)
	}

	if tr.GetCallbacks().AfterFind != nil {
		return tr.GetCallbacks().AfterFind(res)
	}

	return res, nil
}

func (tr *TestRepositoryImpl) prepareFindByCode() error {
	if tr.findByCodeStmt == nil {
		tr.mu.Lock()
		defer tr.mu.Unlock()
		if tr.findByCodeStmt != nil {
			return nil
		}

		var err error = nil

		queryCtx, queryCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer queryCancel()

		tr.findByCodeStmt, err = tr.GetDB().GetDB().PrepareContext(queryCtx, sqlTestFindByCode)
		if err != nil {
			return errs.NewDalError("TestRepository.prepareFindByCode", "prepare find by code stmt", err)
		}
	}

	return nil
}

func (tr *TestRepositoryImpl) entityScanner(scanner repository.Scannable, dest *domain.Test) error {
	return scanner.Scan(&dest.ID, &dest.Code, &dest.Name, &dest.Description, &dest.CreatedAt, &dest.ModifiedAt)
}

func (tr *TestRepositoryImpl) validateCreate(entity *domain.Test) error {
	if entity == nil {
		return errs.NewInvalidArgumentError("entity", "test entity is nil")
	}

	return entity.ValidateCreate()
}

func (tr *TestRepositoryImpl) beforeCreate(entity *domain.Test) error {
	if err := entity.BeforeCreate(); err != nil {
		return errs.NewDalError("TestRepository.beforeCreate", "before create entity", err)
	}

	return nil
}

func (tr *TestRepositoryImpl) creator(ctx context.Context, tx *sql.Tx, entity *domain.Test) (*sql.Row, error) {
	return tx.QueryRowContext(ctx, tr.GetQueryBuilders().GetCreate()(), entity.ID, entity.Code, entity.Name, entity.Description, entity.CreatedAt, entity.ModifiedAt), nil
}

func (tr *TestRepositoryImpl) validateChange(entity *domain.Test) error {
	if entity == nil {
		return errs.NewInvalidArgumentError("entity", "test entity is nil")
	}

	return entity.ValidateChange()
}

func (tr *TestRepositoryImpl) changer(ctx context.Context, tx *sql.Tx, entity *domain.Test) (*sql.Row, error) {
	return tx.QueryRowContext(ctx, tr.GetQueryBuilders().GetChange()(), entity.ID, entity.Code, entity.Name, entity.Description, entity.ModifiedAt), nil
}

func (tr *TestRepositoryImpl) beforeChange(entity *domain.Test) error {
	if err := entity.BeforeChange(); err != nil {
		return errs.NewDalError("TestRepository.beforeChange", "before change entity", err)
	}

	return nil
}

func (tr *TestRepositoryImpl) Close() error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	var errsArr []error
	if tr.findByCodeStmt != nil {
		errsArr = append(errsArr, tr.findByCodeStmt.Close())
	}
	errsArr = append(errsArr, tr.BaseRepository.Close())

	massErrs := errors.Join(errsArr...)
	if massErrs != nil {
		return errs.NewDalError("TestRepositoryImpl.Close", "close resources", massErrs)
	}

	return nil
}
