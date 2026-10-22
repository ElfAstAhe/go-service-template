package v2

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

const (
	StatementFind   string = "statement-find"   // Константа-идентификатор для операции поиска одиночной сущности (Find)
	StatementList   string = "statement-list"   // Константа-идентификатор для операции выборки коллекций данных (List/All)
	StatementCreate string = "statement-create" // Константа-идентификатор для операции вставки записи (Create INSERT)
	StatementChange string = "statement-change" // Константа-идентификатор для операции модификации полей (Change UPDATE)
	StatementDelete string = "statement-delete" // Константа-идентификатор для операции жесткого или мягкого удаления (Delete)
)

type CRUD[E domain.Entity[ID], ID comparable] struct {
	helper *Helper[E, ID]
	opts   *CRUDOptions[E, ID]
}

var _ domain.CRUDRepository[*dummyModel, string] = (*CRUD[*dummyModel, string])(nil)

func NewCRUD[E domain.Entity[ID], ID comparable](options ...CRUDOption[E, ID]) (*CRUD[E, ID], error) {
	// options
	opts := NewCRUDOptions[E, ID]()
	for _, option := range options {
		option(opts)
	}
	if err := opts.Validate(); err != nil {
		return nil, errs.NewDalError("NewCRUD[E domain.Entity[ID]", "options validations failed", err)
	}
	// instance
	res := &CRUD[E, ID]{
		opts: opts,
	}
	// hooks
	hookList := append(opts.Hooks, withHookDispatchMapper[E, ID](res.mapperDispatcher))
	hooks := NewHooks[E, ID]()
	for _, hook := range hookList {
		hook(hooks)
	}
	if err := hooks.Validate(); err != nil {
		return nil, errs.NewDalError("NewCRUD[E domain.Entity[ID]", "hooks validations failed", err)
	}
	// helper
	helper := NewHelper[E, ID](
		opts.Executor,
		opts.ErrorDecipher,
		hooks,
		opts.Statements,
		opts.Info,
	)
	// setup
	res.helper = helper

	return res, nil
}

func (cr *CRUD[E, ID]) Find(ctx context.Context, id ID) (E, error) {
	return cr.helper.Get(ctx, StatementFind, id)
}

func (cr *CRUD[E, ID]) List(ctx context.Context, limit, offset int) ([]E, error) {
	if err := cr.validateList(limit, offset); err != nil {
		return nil, errs.NewDalError("CRUD[E, ID].List", "list failed", err)
	}

	return cr.helper.List(ctx, StatementList, limit, offset)
}

func (cr *CRUD[E, ID]) validateList(limit, offset int) error {
	if limit <= 0 {
		return errs.NewDalError("CRUD[E, ID].validateList", "limit must be greater 0", nil)
	}
	if offset <= 0 {
		return errs.NewDalError("CRUD[E, ID].validateList", "offset must be equal or greater 0", nil)
	}

	return nil
}

func (cr *CRUD[E, ID]) Create(ctx context.Context, entity E) (E, error) {
	if err := cr.validateCreate(ctx, entity); err != nil {
		return cr.helper.GetNilInstance(), errs.NewDalError("CRUD[E, ID].Create", "create entity failed", err)
	}

	return cr.helper.Create(ctx, StatementCreate, entity)
}

func (cr *CRUD[E, ID]) validateCreate(ctx context.Context, entity E) error {
	// check creator
	if utils.IsNil(cr.helper.GetHooks().Creator) {
		return errs.NewNotImplementedError(errs.NewDalError("CRUD[E, ID].validateCreate", "creator hook not applied", nil))
	}
	// validate create
	if !utils.IsNil(cr.helper.GetHooks().ValidateCreate) {
		if err := cr.helper.GetHooks().ValidateCreate(entity); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateCreate", "create entity failed", err)
		}
	}

	return nil
}

func (cr *CRUD[E, ID]) mapperDispatcher(ctx context.Context, stmtName string, scannable Scannable) (E, []any, error) {
	mapper, err := cr.helper.GetStatements().GetMapper(stmtName)
	if err != nil {
		return cr.helper.GetNilInstance(), nil, errs.NewDalError("CRUD[E, ID].mapperDispatcher", fmt.Sprintf("mapper %s not found", stmtName), err)
	}

	return mapper(ctx, scannable)
}
