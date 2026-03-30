package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type Helper[T domain.Entity[ID], ID comparable] struct {
	exec        db.Executor
	errDecipher db.ErrorDecipher
	info        *EntityInfo
	nilInstance T
	callbacks   *BaseRepositoryCallbacks[T, ID]
}

func newHelper[T domain.Entity[ID], ID comparable](exec db.Executor, errDecipher db.ErrorDecipher, callbacks *BaseRepositoryCallbacks[T, ID], info *EntityInfo) *Helper[T, ID] {
	return &Helper[T, ID]{
		exec:        exec,
		errDecipher: errDecipher,
		info:        info,
		callbacks:   callbacks,
	}
}

func (h *Helper[T, ID]) GetExecutor() db.Executor {
	return h.exec
}

func (h *Helper[T, ID]) GetErrDecipher() db.ErrorDecipher {
	return h.errDecipher
}

func (h *Helper[T, ID]) GetInfo() *EntityInfo {
	return h.info
}

func (h *Helper[T, ID]) GetNilInstance() T {
	return h.nilInstance
}

func (h *Helper[T, ID]) GetCallbacks() *BaseRepositoryCallbacks[T, ID] {
	return h.callbacks
}

func (h *Helper[T, ID]) Get(ctx context.Context, sourceLabel string, sqlReq string, params ...any) (T, error) {
	// Получаем querier (либо транзакция, либо БД)
	querier := h.exec.GetQuerier(ctx)

	row := querier.QueryRowContext(ctx, sqlReq, params...)
	res := h.callbacks.NewEntityFactory()

	err := h.callbacks.EntityScanner(row, sourceLabel, res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return h.nilInstance, errs.NewDalNotFoundError(h.info.Entity, params, err)
		}

		return h.nilInstance, errs.NewDalError("Helper.Get", "get row", err)
	}

	if h.callbacks.AfterFind != nil {
		return h.callbacks.AfterFind(res, params...)
	}

	return res, nil
}

func (h *Helper[T, ID]) List(ctx context.Context, sourceLabel string, sqlReq string, params ...any) ([]T, error) {
	querier := h.GetExecutor().GetQuerier(ctx)

	rows, err := querier.QueryContext(ctx, sqlReq, params...)
	if err != nil {
		return nil, errs.NewDalError("Helper.List", "query", err)
	}
	defer rows.Close()

	res := make([]T, 0)
	for rows.Next() {
		if err = ctx.Err(); err != nil {
			return nil, errs.NewDalError("Helper.List", "check context", err)
		}

		addEntity := true
		entity := h.GetCallbacks().NewEntityFactory()

		err = h.GetCallbacks().EntityScanner(rows, sourceLabel, entity, params...)
		if err != nil {
			return nil, errs.NewDalError("Helper.List", "scan rows", err)
		}

		if h.GetCallbacks().AfterListYield != nil {
			entity, addEntity, err = h.GetCallbacks().AfterListYield(entity, params...)
			if err != nil {
				return nil, errs.NewDalError("Helper.List", "post scan processing", err)
			}
		}
		// yeld метод постобработки строки не вернул entity, нет данных - нет добавления
		if any(entity) == nil || !addEntity {
			continue
		}

		res = append(res, entity)
	}
	if rows.Err() != nil {
		return nil, errs.NewDalError("Helper.List", "after scan", rows.Err())
	}

	return res, nil
}

func (h *Helper[T, ID]) Create(ctx context.Context, sourceLabel string, entity T, params ...any) (T, error) {
	if h.GetCallbacks().BeforeCreate != nil {
		if err := h.GetCallbacks().BeforeCreate(entity, params...); err != nil {
			return h.GetNilInstance(), errs.NewDalError("Helper.Create", "before create", err)
		}
	}

	querier := h.GetExecutor().GetQuerier(ctx)

	row, err := h.GetCallbacks().Creator(ctx, querier, entity, params...)
	if err != nil {
		return h.GetNilInstance(), errs.NewDalError("Helper.Create", "create entity", err)
	}

	res := h.GetCallbacks().NewEntityFactory()
	err = h.GetCallbacks().EntityScanner(row, sourceLabel, res, params...)
	if err != nil {
		if h.errDecipher.IsUniqueViolation(err) {
			return h.GetNilInstance(), errs.NewDalAlreadyExistsError(h.GetInfo().Entity, entity.GetID(), err)
		}

		return h.GetNilInstance(), errs.NewDalError("Helper.Create", "scan after create entity", err)
	}

	if h.GetCallbacks().AfterFind != nil {
		return h.GetCallbacks().AfterFind(res, params...)
	}

	return res, nil
}

func (h *Helper[T, ID]) Change(ctx context.Context, sourceLabel string, entity T, params ...any) (T, error) {
	if h.GetCallbacks().BeforeChange != nil {
		if err := h.GetCallbacks().BeforeChange(entity, params...); err != nil {
			return h.GetNilInstance(), errs.NewDalError("Helper.Change", "before change", err)
		}
	}

	querier := h.GetExecutor().GetQuerier(ctx)

	row, err := h.GetCallbacks().Changer(ctx, querier, entity, params...)
	if err != nil {
		return h.GetNilInstance(), errs.NewDalError("Helper.Change", "change entity", err)
	}

	res := h.GetCallbacks().NewEntityFactory()
	err = h.GetCallbacks().EntityScanner(row, sourceLabel, res, params...)
	if err != nil {
		if h.errDecipher.IsUniqueViolation(err) {
			return h.GetNilInstance(), errs.NewDalAlreadyExistsError(h.GetInfo().Entity, entity.GetID(), err)
		}

		return h.GetNilInstance(), errs.NewDalError("Helper.Change", "scan after change entity", err)
	}

	if h.GetCallbacks().AfterFind != nil {
		return h.GetCallbacks().AfterFind(res, params...)
	}

	return res, nil
}

func (h *Helper[T, ID]) Delete(ctx context.Context, sqlReq string, params ...any) error {
	// Получаем querier (либо транзакция, либо БД)
	querier := h.GetExecutor().GetQuerier(ctx)
	res, err := querier.ExecContext(ctx, sqlReq, params...)
	if err != nil {
		return errs.NewDalError("Helper.Delete", "exec context", err)
	}
	// проверяем
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errs.NewDalError("Helper.Delete", "rows affected", err)
	}
	if !(rowsAffected > 0) {
		return errs.NewDalNotFoundError(h.GetInfo().Entity, params, nil)
	}

	return nil
}

func (h *Helper[T, ID]) DeleteNoCheck(ctx context.Context, sqlReq string, params ...any) error {
	// Получаем querier (либо транзакция, либо БД)
	querier := h.GetExecutor().GetQuerier(ctx)
	_, err := querier.ExecContext(ctx, sqlReq, params...)
	if err != nil {
		return errs.NewDalError("Helper.DeleteNoCheck", "exec context", err)
	}

	return nil
}
