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
		opts.TableName,
		opts.EntityName,
	)
	// setup
	res.helper = helper

	return res, nil
}

// Find выполняет поиск и извлечение одиночной доменной сущности по её уникальному идентификатору.
func (cr *CRUD[E, ID]) Find(ctx context.Context, id ID) (E, error) {
	if err := cr.validateFind(ctx, id); err != nil {
		return cr.helper.GetNilInstance(), errs.NewDalError("CRUD[E, ID].Find", "find failed", err)
	}

	return cr.helper.Get(ctx, StatementFind, id)
}

func (cr *CRUD[E, ID]) validateFind(ctx context.Context, id ID) error {
	if !utils.IsNil(cr.helper.GetHooks().ValidateFind) {
		if err := cr.helper.GetHooks().ValidateFind(ctx, id); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateFind", "find validate hook failed", err)
		}
	}

	return nil
}

// List возвращает постраничную выборку (срез) доменных сущностей с учетом лимитов и смещений.
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
	if offset < 0 {
		return errs.NewDalError("CRUD[E, ID].validateList", "offset must be equal or greater 0", nil)
	}

	return nil
}

// Create запускает валидацию бизнес-инвариантов и атомарно вставляет новую сущность в базу данных.
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
	// domain validation
	if err := entity.ValidateCreate(); err != nil {
		return errs.NewDalError("CRUD[E, ID].validateCreate", "entity validation failed", err)
	}
	// validate create
	if !utils.IsNil(cr.helper.GetHooks().ValidateCreate) {
		if err := cr.helper.GetHooks().ValidateCreate(ctx, entity); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateCreate", "create validate hook failed", err)
		}
	}

	return nil
}

// Change выполняет обновление (UPDATE) полей существующей сущности.
func (cr *CRUD[E, ID]) Change(ctx context.Context, entity E) (E, error) {
	if err := cr.validateChange(ctx, entity); err != nil {
		return cr.helper.GetNilInstance(), errs.NewDalError("CRUD[E, ID].Change", "change entity failed", err)
	}

	return cr.helper.Change(ctx, StatementChange, entity)
}

func (cr *CRUD[E, ID]) validateChange(ctx context.Context, entity E) error {
	// check changer
	if utils.IsNil(cr.helper.GetHooks().Changer) {
		return errs.NewNotImplementedError(errs.NewDalError("CRUD[E, ID].validateChange", "changer hook not applied", nil))
	}
	// domain validation
	if err := entity.ValidateChange(); err != nil {
		return errs.NewDalError("CRUD[E, ID].validateChange", "entity validation failed", err)
	}
	// validate change
	if !utils.IsNil(cr.helper.GetHooks().ValidateChange) {
		if err := cr.helper.GetHooks().ValidateChange(ctx, entity); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateChange", "change validate hook failed", err)
		}
	}

	return nil
}

// Delete удаляет запись из физической таблицы по её уникальному идентификатору ID.
func (cr *CRUD[E, ID]) Delete(ctx context.Context, id ID) error {
	if err := cr.validateDelete(ctx, id); err != nil {
		return errs.NewDalError("CRUD[E, ID].Delete", "delete failed", err)
	}

	return cr.helper.Delete(ctx, StatementDelete, id)
}

func (cr *CRUD[E, ID]) validateDelete(ctx context.Context, id ID) error {
	if !utils.IsNil(cr.helper.GetHooks().ValidateDelete) {
		if err := cr.helper.GetHooks().ValidateDelete(ctx, id); err != nil {
			return errs.NewDalError("CRUD[E, ID].validateDelete", "delete validate hook failed", err)
		}
	}

	return nil
}

func (cr *CRUD[E, ID]) mapperDispatcher(ctx context.Context, stmtName string, scannable Scannable, params ...any) (E, error) {
	mapper, err := cr.helper.GetStatements().GetMapper(stmtName)
	if err != nil {
		return cr.helper.GetNilInstance(), errs.NewDalError("CRUD[E, ID].mapperDispatcher", fmt.Sprintf("mapper %s not found", stmtName), err)
	}

	return mapper(ctx, scannable, params...)
}
