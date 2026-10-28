package repository

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	pkgdbpg "github.com/ElfAstAhe/go-service-template/pkg/db/postgres"
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkgrepo "github.com/ElfAstAhe/go-service-template/pkg/repository/v2"
)

const statementTest2FindByCode string = "statement-test2-find-by-code"

// Test2 crud репозиторий сущности domain.Test2
type Test2 struct {
	*pkgrepo.CRUD[*domain.Test2, string]
}

var _ pkgdom.CRUDRepository[*domain.Test2, string] = (*Test2)(nil)
var _ domain.Test2Repository = (*Test2)(nil)

var test2StmtBuilders map[string]func(*Test2) []pkgrepo.CRUDOption[*domain.Test2, string]

// NewTest2 фабричная функция создания экземпляра репозитория
func NewTest2(
	driverName string,
	executor db.Executor,
	errDecipher db.ErrorDecipher,
) (*Test2, error) {
	// instance
	res := &Test2{}
	// statements
	stmtOpts, err := buildTest2StmtOptions(driverName, res)
	if err != nil {
		return nil, errs.NewDalError("NewTest2", "build statements failed", err)
	}
	// options
	opts := []pkgrepo.CRUDOption[*domain.Test2, string]{
		pkgrepo.WithExecutor[*domain.Test2, string](executor),
		pkgrepo.WithErrorDecipher[*domain.Test2, string](errDecipher),
		pkgrepo.WithHooks[*domain.Test2, string](
			pkgrepo.WithHookNewEntityFactory[*domain.Test2, string](domain.NewEmptyTest2),
			pkgrepo.WithHookCreator[*domain.Test2, string](res.creator),
			pkgrepo.WithHookChanger[*domain.Test2, string](res.changer),
		),
		pkgrepo.WithEntityInfo[*domain.Test2, string]("test2", "Test2"),
	}
	opts = append(opts, stmtOpts...)
	// embed
	embed, err := pkgrepo.NewCRUD[*domain.Test2, string](opts...)
	if err != nil {
		return nil, errs.NewDalError("NewTest2", "create embed CRUD failed", err)
	}
	// setup
	res.CRUD = embed

	return res, nil
}

// FindByCode осуществляет поиск сущности по коду
func (t2 *Test2) FindByCode(ctx context.Context, code string) (*domain.Test2, error) {
	return t2.GetHelper().Get(ctx, statementTest2FindByCode, code)
}

func (t2 *Test2) creator(ctx context.Context, querier db.Querier, entity *domain.Test2, params ...any) (pkgrepo.Scannable, error) {
	query, err := t2.GetHelper().GetStatements().GetSQL(pkgrepo.StatementCreate)
	if err != nil {
		return nil, errs.NewDalError("Test2", "create statements not found", err)
	}

	return querier.QueryRowContext(ctx, query,
		entity.ID,
		entity.Code,
		entity.Name,
		entity.Description,
		entity.CreatedAt,
		entity.ModifiedAt,
	), nil
}

func (t2 *Test2) changer(ctx context.Context, querier db.Querier, entity *domain.Test2, params ...any) (pkgrepo.Scannable, error) {
	query, err := t2.GetHelper().GetStatements().GetSQL(pkgrepo.StatementChange)
	if err != nil {
		return nil, errs.NewDalError("Test2", "change statements not found", err)
	}

	return querier.QueryRowContext(ctx, query,
		entity.ID,
		entity.Code,
		entity.Name,
		entity.Description,
		entity.ModifiedAt,
	), nil
}

func (t2 *Test2) mapDefault(ctx context.Context, scannable pkgrepo.Scannable, params ...any) (*domain.Test2, error) {
	res := t2.GetHelper().GetHooks().NewEntityFactory()
	err := scannable.Scan(
		&res.ID,
		&res.Code,
		&res.Name,
		&res.Description,
		&res.CreatedAt,
		&res.ModifiedAt,
	)
	if err != nil {
		return nil, errs.NewDalError("Test2.mapDefault", "default mapper failed", err)
	}

	return res, nil
}

func (t2 *Test2) pgStatementsBuilder() []pkgrepo.CRUDOption[*domain.Test2, string] {
	return []pkgrepo.CRUDOption[*domain.Test2, string]{
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementFind, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2Find, t2.mapDefault)),
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementList, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2List, t2.mapDefault)),
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementCreate, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2Create, t2.mapDefault)),
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementChange, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2Change, t2.mapDefault)),
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementDelete, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2Delete, nil)),
		pkgrepo.WithStatement[*domain.Test2, string](statementTest2FindByCode, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2FindByCode, t2.mapDefault)),
	}
}

// =======================================================
// Statements
// =======================================================

func initTest2StmtBuilderOptions() {
	test2StmtBuilders = map[string]func(*Test2) []pkgrepo.CRUDOption[*domain.Test2, string]{
		pkgdbpg.DriverName: buildTest2PgStmtOptions,
		"pgx":              buildTest2PgStmtOptions,
		"postgres":         buildTest2PgStmtOptions,
	}
}

func buildTest2StmtOptions(driver string, inst *Test2) ([]pkgrepo.CRUDOption[*domain.Test2, string], error) {
	builder, ok := test2StmtBuilders[driver]
	if !ok {
		return nil, errs.NewDalError("buildStatements", fmt.Sprintf("unknown driver [%s]", driver), nil)
	}

	return builder(inst), nil
}

func buildTest2PgStmtOptions(inst *Test2) []pkgrepo.CRUDOption[*domain.Test2, string] {
	return []pkgrepo.CRUDOption[*domain.Test2, string]{
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementFind, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2Find, inst.mapDefault)),
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementList, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2List, inst.mapDefault)),
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementCreate, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2Create, inst.mapDefault)),
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementChange, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2Change, inst.mapDefault)),
		pkgrepo.WithStatement[*domain.Test2, string](pkgrepo.StatementDelete, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2Delete, nil)),
		pkgrepo.WithStatement[*domain.Test2, string](statementTest2FindByCode, pkgrepo.NewStatement[*domain.Test2, string](postgres.SqlTest2FindByCode, inst.mapDefault)),
	}
}

func init() {
	initTest2StmtBuilderOptions()
}
