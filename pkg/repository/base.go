package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// BaseRepository базовая реализация CRUD репозитория
type BaseRepository[T domain.Entity[ID], ID any] struct {
	exec        db.Executor
	errDecipher db.ErrorDecipher
	tm          db.TransactionManager
	info        *EntityInfo

	nilInstance T

	queryBuilders *BaseQueryBuilders
	callbacks     *BaseRepositoryCallbacks[T, ID]
}

func NewBaseRepository[T domain.Entity[ID], ID any](
	exec db.Executor,
	errDecipher db.ErrorDecipher,
	info *EntityInfo,
	queryBuilders *BaseQueryBuilders,
	callbacks *BaseRepositoryCallbacks[T, ID],
) (*BaseRepository[T, ID], error) {
	return &BaseRepository[T, ID]{
		exec:          exec,
		errDecipher:   errDecipher,
		info:          info,
		queryBuilders: queryBuilders,
		callbacks:     callbacks,
	}, nil
}

func (br *BaseRepository[T, ID]) Find(ctx context.Context, id ID) (T, error) {
	// Получаем querier (либо транзакция, либо БД)
	querier := br.exec.GetQuerier(ctx)

	sqlFind, err := br.prepareFind()
	if err != nil {
		return br.nilInstance, err
	}

	row := querier.QueryRowContext(ctx, sqlFind, id)
	res := br.callbacks.NewEntityFactory()

	// res, err := br.callbacks.RowScanner(br.findStmt.QueryRowContext(ctx, id))
	err = br.callbacks.EntityScanner(row, res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return br.nilInstance, errs.NewDalNotFoundError(br.info.Entity, id, err)
		}

		return br.nilInstance, errs.NewDalError("BaseRepository.Find", "get row", err)
	}

	if br.callbacks.AfterFind != nil {
		return br.callbacks.AfterFind(res)
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) prepareFind() (string, error) {
	if br.queryBuilders == nil {
		return "", errs.NewDalError("BaseRepository.prepareFind", "query builders not applied", nil)
	}
	if br.queryBuilders.findBuilder == nil {
		return "", errs.NewDalError("BaseRepository.prepareFind", "query find builder not applied", nil)
	}
	sqlFind := br.queryBuilders.findBuilder()
	if strings.TrimSpace(sqlFind) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseRepository.prepareFind", "query find empty", nil))
	}

	return sqlFind, nil
}

func (br *BaseRepository[T, ID]) List(ctx context.Context, limit, offset int) ([]T, error) {
	if err := br.ValidateList(limit, offset); err != nil {
		return nil, errs.NewDalError("BaseRepository.List", "validate list", err)
	}
	sqlList, err := br.prepareList()
	if err != nil {
		return nil, errs.NewNotImplementedError(err)
	}

	return br.InternalList(ctx, sqlList, limit, offset)
}

func (br *BaseRepository[T, ID]) ValidateList(limit, offset int) error {
	if !(limit > 0) {
		return errs.NewDalError("BaseRepository.ValidateList", "limit must be greater 0", nil)
	}
	if !(offset >= 0) {
		return errs.NewDalError("BaseRepository.ValidateList", "offset must be equal or greater 0", nil)
	}

	return nil
}

func (br *BaseRepository[T, ID]) prepareList() (string, error) {
	sqlList := br.queryBuilders.listBuilder()
	if strings.TrimSpace(sqlList) == "" {
		return "", errs.NewDalError("BaseRepository.prepareList", "sql list empty", nil)
	}

	return sqlList, nil
}

//goland:noinspection GoUnhandledErrorResult
func (br *BaseRepository[T, ID]) InternalList(ctx context.Context, sqlReq string, params ...any) ([]T, error) {
	querier := br.exec.GetQuerier(ctx)

	rows, err := querier.QueryContext(ctx, sqlReq, params...)
	if err != nil {
		return nil, errs.NewDalError("BaseRepository.List", "query", err)
	}
	defer rows.Close()

	res := make([]T, 0)
	for rows.Next() {
		if err = ctx.Err(); err != nil {
			return nil, errs.NewDalError("BaseRepository.InternalList", "check context", err)
		}

		addEntity := true
		entity := br.callbacks.NewEntityFactory()

		err = br.callbacks.EntityScanner(rows, entity)
		if err != nil {
			return nil, errs.NewDalError("BaseRepository.InternalList", "scan rows", err)
		}

		if br.callbacks.AfterListYield != nil {
			entity, addEntity, err = br.callbacks.AfterListYield(entity)
			if err != nil {
				return nil, errs.NewDalError("BaseRepository.InternalList", "post scan processing", err)
			}
		}
		// yeld метод постобработки строки не вернул entity, нет данных - нет добавления
		if any(entity) == nil || !addEntity {
			continue
		}

		res = append(res, entity)
	}
	if rows.Err() != nil {
		return nil, errs.NewDalError("BaseRepository.InternalList", "after scan", rows.Err())
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) Create(ctx context.Context, entity T) (T, error) {
	if br.callbacks.Creator == nil {
		return br.nilInstance, errs.NewNotImplementedError(errs.NewDalError("BaseRepository.Create", "creator not applied", nil))
	}

	if br.callbacks.ValidateCreate != nil {
		if err := br.callbacks.ValidateCreate(entity); err != nil {
			return br.nilInstance, errs.NewDalError("BaseRepository.Create", "validate create", err)
		}
	}

	if br.callbacks.BeforeCreate != nil {
		if err := br.callbacks.BeforeCreate(entity); err != nil {
			return br.nilInstance, errs.NewDalError("BaseRepository.Create", "before create", err)
		}
	}

	querier := br.exec.GetQuerier(ctx)

	row, err := br.callbacks.Creator(ctx, querier, entity)
	if err != nil {
		return br.nilInstance, errs.NewDalError("BaseRepository.Create", "create entity", err)
	}

	res := br.callbacks.NewEntityFactory()
	err = br.callbacks.EntityScanner(row, res)
	if err != nil {
		if br.errDecipher.IsUniqueViolation(err) {
			return br.nilInstance, errs.NewDalAlreadyExistsError(br.info.Entity, nil, err)
		}

		return br.nilInstance, errs.NewDalError("BaseRepository.Create", "scan after create entity", err)
	}

	if br.callbacks.AfterFind != nil {
		return br.callbacks.AfterFind(res)
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) Change(ctx context.Context, entity T) (T, error) {
	if br.callbacks.Changer == nil {
		return br.nilInstance, errs.NewNotImplementedError(errs.NewDalError("BaseRepository.Change", "changer not applied", nil))
	}

	if br.callbacks.ValidateChange != nil {
		if err := br.callbacks.ValidateChange(entity); err != nil {
			return br.nilInstance, errs.NewDalError("BaseRepository.Change", "validate change", err)
		}
	}

	if br.callbacks.BeforeChange != nil {
		if err := br.callbacks.BeforeChange(entity); err != nil {
			return br.nilInstance, errs.NewDalError("BaseRepository.Change", "before change", err)
		}
	}

	querier := br.exec.GetQuerier(ctx)

	row, err := br.callbacks.Changer(ctx, querier, entity)
	if err != nil {
		return br.nilInstance, errs.NewDalError("BaseRepository.Change", "change entity", err)
	}

	res := br.callbacks.NewEntityFactory()
	err = br.callbacks.EntityScanner(row, res)
	if err != nil {
		if br.errDecipher.IsUniqueViolation(err) {
			return br.nilInstance, errs.NewDalAlreadyExistsError(br.info.Entity, entity.GetID(), err)
		}

		return br.nilInstance, errs.NewDalError("BaseRepository.Change", "scan after change entity", err)
	}

	if br.callbacks.AfterFind != nil {
		return br.callbacks.AfterFind(res)
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	// Получаем querier (либо транзакция, либо БД)
	querier := br.exec.GetQuerier(ctx)

	sqlDelete, err := br.prepareDelete()
	if err != nil {
		return errs.NewNotImplementedError(err)
	}

	res, err := querier.ExecContext(ctx, sqlDelete, id)
	if err != nil {
		return errs.NewDalError("BaseRepository.Delete", "exec context", err)
	}
	// проверяем
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errs.NewDalError("BaseRepository.Delete", "rows affected", err)
	}
	if !(rowsAffected > 0) {
		return errs.NewDalNotFoundError(br.info.Entity, id, nil)
	}

	return nil
}

func (br *BaseRepository[T, ID]) prepareDelete() (string, error) {
	if br.queryBuilders == nil {
		return "", errs.NewDalError("BaseRepository.prepareDelete", "query builders not applied", nil)
	}
	if br.queryBuilders.deleteBuilder == nil {
		return "", errs.NewDalError("BaseRepository.prepareDelete", "query delete builder not applied", nil)
	}
	sqlDelete := br.queryBuilders.deleteBuilder()
	if strings.TrimSpace(sqlDelete) == "" {
		return "", errs.NewDalError("BaseRepository.prepareDelete", "query delete empty", nil)
	}

	return sqlDelete, nil
}

func (br *BaseRepository[T, ID]) Close() error {
	return nil
}

func (br *BaseRepository[T, ID]) GetInfo() *EntityInfo {
	return br.info
}

func (br *BaseRepository[T, ID]) GetExecutor() db.Executor {
	return br.exec
}

func (br *BaseRepository[T, ID]) GetErrDecipher() db.ErrorDecipher {
	return br.errDecipher
}

func (br *BaseRepository[T, ID]) GetQueryBuilders() *BaseQueryBuilders {
	return br.queryBuilders
}

func (br *BaseRepository[T, ID]) GetCallbacks() *BaseRepositoryCallbacks[T, ID] {
	return br.callbacks
}
