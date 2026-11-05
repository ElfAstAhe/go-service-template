package postgres

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
)

type TestRepository struct {
	*repository.CRUD[*domain.Test, string]
}

var _ pkgdom.CRUDRepository[*domain.Test, string] = (*TestRepository)(nil)
var _ domain.TestRepository = (*TestRepository)(nil)

func NewTestRepository(executor db.Executor, decipher db.ErrorDecipher) (*TestRepository, error) {
	// new instance
	res := &TestRepository{}
	// sql builders
	queryBuilders := repository.NewBaseCRUDQueryBuildersBuilder().NewInstance().
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
	callbacks, err := repository.NewCallbacksBuilder[*domain.Test, string]().NewInstance().
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
	base, err := repository.NewCRUD[*domain.Test, string](
		executor,
		decipher,
		repository.NewEntityInfo("test", "Test"),
		queryBuilders,
		callbacks,
	)
	if err != nil {
		return nil, errs.NewCommonError("error create TestRepository", err)
	}

	res.CRUD = base

	return res, nil
}

func (tr *TestRepository) FindByCode(ctx context.Context, code string) (*domain.Test, error) {
	return tr.GetHelper().Get(ctx, sqlTestFindByCode, code)
}

func (tr *TestRepository) entityScanner(scanner repository.Scannable, sourceLabel string, dest *domain.Test, params ...any) error {
	return scanner.Scan(&dest.ID, &dest.Code, &dest.Name, &dest.Description, &dest.CreatedAt, &dest.ModifiedAt)
}

func (tr *TestRepository) validateCreate(entity *domain.Test, params ...any) error {
	if entity == nil {
		return errs.NewInvalidArgumentError("entity", "test entity is nil")
	}

	return entity.ValidateCreate()
}

func (tr *TestRepository) beforeCreate(entity *domain.Test, params ...any) error {
	if err := entity.BeforeCreate(); err != nil {
		return errs.NewDalError("TestRepository.beforeCreate", "before create entity", err)
	}

	return nil
}

func (tr *TestRepository) creator(ctx context.Context, querier db.Querier, entity *domain.Test, params ...any) (*sql.Row, error) {
	return querier.QueryRowContext(ctx, tr.GetQueryBuilders().GetCreate()(), entity.ID, entity.Code, entity.Name, entity.Description, entity.CreatedAt, entity.ModifiedAt), nil
}

func (tr *TestRepository) validateChange(entity *domain.Test, params ...any) error {
	if entity == nil {
		return errs.NewInvalidArgumentError("entity", "test entity is nil")
	}

	return entity.ValidateChange()
}

func (tr *TestRepository) changer(ctx context.Context, querier db.Querier, entity *domain.Test, params ...any) (*sql.Row, error) {
	return querier.QueryRowContext(ctx, tr.GetQueryBuilders().GetChange()(), entity.ID, entity.Code, entity.Name, entity.Description, entity.ModifiedAt), nil
}

func (tr *TestRepository) beforeChange(entity *domain.Test, params ...any) error {
	if err := entity.BeforeChange(); err != nil {
		return errs.NewDalError("TestRepository.beforeChange", "before change entity", err)
	}

	return nil
}
