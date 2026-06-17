package postgres

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
)

type TestRepositoryImpl struct {
	*repository.BaseCRUDRepository[*domain.Test, string]
}

var _ domain.TestRepository = (*TestRepositoryImpl)(nil)

func NewTestRepository(executor db.Executor, decipher db.ErrorDecipher) (*TestRepositoryImpl, error) {
	// new instance
	res := &TestRepositoryImpl{}
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
	base, err := repository.NewBaseCRUDRepository[*domain.Test, string](
		executor,
		decipher,
		repository.NewEntityInfo("test", "Test"),
		queryBuilders,
		callbacks,
	)
	if err != nil {
		return nil, errs.NewCommonError("error create TestRepository", err)
	}

	res.BaseCRUDRepository = base

	return res, nil
}

func (tr *TestRepositoryImpl) FindByCode(ctx context.Context, code string) (*domain.Test, error) {
	return tr.GetHelper().Get(ctx, sqlTestFindByCode, code)
}

func (tr *TestRepositoryImpl) entityScanner(scanner repository.Scannable, sourceLabel string, dest *domain.Test, params ...any) error {
	return scanner.Scan(&dest.ID, &dest.Code, &dest.Name, &dest.Description, &dest.CreatedAt, &dest.ModifiedAt)
}

func (tr *TestRepositoryImpl) validateCreate(entity *domain.Test, params ...any) error {
	if entity == nil {
		return errs.NewInvalidArgumentError("entity", "test entity is nil")
	}

	return entity.ValidateCreate()
}

func (tr *TestRepositoryImpl) beforeCreate(entity *domain.Test, params ...any) error {
	if err := entity.BeforeCreate(); err != nil {
		return errs.NewDalError("TestRepository.beforeCreate", "before create entity", err)
	}

	return nil
}

func (tr *TestRepositoryImpl) creator(ctx context.Context, querier db.Querier, entity *domain.Test, params ...any) (*sql.Row, error) {
	return querier.QueryRowContext(ctx, tr.GetQueryBuilders().GetCreate()(), entity.ID, entity.Code, entity.Name, entity.Description, entity.CreatedAt, entity.ModifiedAt), nil
}

func (tr *TestRepositoryImpl) validateChange(entity *domain.Test, params ...any) error {
	if entity == nil {
		return errs.NewInvalidArgumentError("entity", "test entity is nil")
	}

	return entity.ValidateChange()
}

func (tr *TestRepositoryImpl) changer(ctx context.Context, querier db.Querier, entity *domain.Test, params ...any) (*sql.Row, error) {
	return querier.QueryRowContext(ctx, tr.GetQueryBuilders().GetChange()(), entity.ID, entity.Code, entity.Name, entity.Description, entity.ModifiedAt), nil
}

func (tr *TestRepositoryImpl) beforeChange(entity *domain.Test, params ...any) error {
	if err := entity.BeforeChange(); err != nil {
		return errs.NewDalError("TestRepository.beforeChange", "before change entity", err)
	}

	return nil
}
