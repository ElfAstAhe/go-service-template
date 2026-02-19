package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/helper"
)

// BaseRepository базовая реализация CRUD репозитория
type BaseRepository[T domain.Identity[ID], ID any] struct {
	onceFind   sync.Once
	onceDelete sync.Once
	db         db.DB
	info       *EntityInfo

	findStmt   *sql.Stmt
	deleteStmt *sql.Stmt

	queryBuilders *BaseQueryBuilders
	callbacks     *BaseRepositoryCallbacks[T, ID]
}

//goland:noinspection GoResourceLeak
func NewBaseRepository[T domain.Identity[ID], ID any](
	db db.DB,
	info *EntityInfo,
	queryBuilders *BaseQueryBuilders,
	callbacks *BaseRepositoryCallbacks[T, ID],
) (*BaseRepository[T, ID], error) {
	return &BaseRepository[T, ID]{
		db:            db,
		info:          info,
		findStmt:      nil,
		deleteStmt:    nil,
		queryBuilders: queryBuilders,
		callbacks:     callbacks,
	}, nil
}

func (br *BaseRepository[T, ID]) Find(ctx context.Context, id ID) (*T, error) {
	if err := br.prepareFind(); err != nil {
		return nil, errs.NewNotImplementedError(err)
	}

	res, err := br.callbacks.RowScanner(br.findStmt.QueryRowContext(ctx, id))
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return nil, errs.NewDalNotFoundError(br.info.Entity, id, err)
		}

		return nil, errs.NewDalError("BaseRepository.Find", "get row", err)
	}

	if br.callbacks.AfterFind != nil {
		return br.callbacks.AfterFind(res)
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) prepareFind() error {
	var err error = nil
	br.onceFind.Do(func() {
		if br.queryBuilders == nil {
			err = errs.NewDalError("BaseRepository.Find", "query builders not applied", nil)
		}
		if br.queryBuilders.findBuilder == nil {
			err = errs.NewDalError("BaseRepository.Find", "query find builder not applied", nil)
		}
		sqlFind := br.queryBuilders.findBuilder()
		if strings.TrimSpace(sqlFind) == "" {
			err = errs.NewDalError("BaseRepository.Find", "query find empty", nil)
		}

		queryCtx, queryCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer queryCancel()

		br.findStmt, err = br.db.GetDB().PrepareContext(queryCtx, br.queryBuilders.GetFind()())
		if err != nil {
			err = errs.NewDalError("BaseRepository.Find", "prepare find stmt", err)
		}
	})
	if err != nil {
		return errs.NewDalError("BaseRepository.Find", "prepare find ", err)
	}

	return nil
}

func (br *BaseRepository[T, ID]) List(ctx context.Context, limit, offset int) ([]*T, error) {
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

func (br *BaseRepository[T, ID]) InternalList(ctx context.Context, sqlReq string, params ...any) ([]*T, error) {
	rows, err := br.db.GetDB().QueryContext(ctx, sqlReq, params...)
	if err != nil {
		return nil, errs.NewDalError("BaseRepository.List", "query", err)
	}
	defer rows.Close()

	res := make([]*T, 0)
	for rows.Next() {
		if err = ctx.Err(); err != nil {
			return nil, errs.NewDalError("BaseRepository.InternalList", "check context", err)
		}

		entity := br.callbacks.NewEntityFactory()

		err = br.callbacks.RowsScanner(ctx, rows, entity)
		if err != nil {
			return nil, errs.NewDalError("BaseRepository.InternalList", "scan rows", err)
		}

		if br.callbacks.AfterListYield != nil {
			entity, err = br.callbacks.AfterListYield(entity)
			if err != nil {
				return nil, errs.NewDalError("BaseRepository.InternalList", "post scan processing", err)
			}
		}
		// yeld метод постобработки строки не вернул entity, нет данных - нет добавления
		if entity == nil {
			continue
		}

		res = append(res, entity)
	}
	if rows.Err() != nil {
		return nil, errs.NewDalError("BaseRepository.InternalList", "after scan", rows.Err())
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) Create(ctx context.Context, entity *T) (*T, error) {
	if br.callbacks.Creator == nil {
		return nil, errs.NewNotImplementedError(errs.NewDalError("BaseRepository.Create", "creator not applied", nil))
	}

	if br.callbacks.ValidateCreate != nil {
		if err := br.callbacks.ValidateCreate(entity); err != nil {
			return nil, errs.NewDalError("BaseRepository.Create", "validate create", err)
		}
	}

	if br.callbacks.BeforeCreate != nil {
		if err := br.callbacks.BeforeCreate(entity); err != nil {
			return nil, errs.NewDalError("BaseRepository.Create", "before create", err)
		}
	}

	res := br.callbacks.NewEntityFactory()
	// выполнение
	err := br.db.GetHelper().RunInTx(ctx, br.db.GetDB(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		row, err := br.callbacks.Creator(ctx, tx, entity)
		if err != nil {
			return errs.NewDalError("BaseRepository.Create", "create entity", err)
		}

		res, err = br.callbacks.RowScanner(row)
		if err != nil {
			if br.db.GetHelper().IsUniqueViolation(err) {
				return errs.NewDalAlreadyExistsError(br.info.Entity, nil, err)
			}

			return errs.NewDalError("BaseRepository.Create", "scan after create entity", err)
		}

		return nil
	})
	if err != nil {
		return nil, errs.NewDalError("BaseRepository.Create", "run in transaction", err)
	}

	if br.callbacks.AfterFind != nil {
		return br.callbacks.AfterFind(res)
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) Change(ctx context.Context, entity *T) (*T, error) {
	if br.callbacks.Changer == nil {
		return nil, errs.NewNotImplementedError(errs.NewDalError("BaseRepository.Change", "changer not applied", nil))
	}

	if br.callbacks.ValidateChange != nil {
		if err := br.callbacks.ValidateChange(entity); err != nil {
			return nil, errs.NewDalError("BaseRepository.Change", "validate change", err)
		}
	}

	if br.callbacks.BeforeChange != nil {
		if err := br.callbacks.BeforeChange(entity); err != nil {
			return nil, errs.NewDalError("BaseRepository.Change", "before change", err)
		}
	}

	res := br.callbacks.NewEntityFactory()
	// выполнение
	err := br.db.GetHelper().RunInTx(ctx, br.db.GetDB(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		row, err := br.callbacks.Changer(ctx, tx, entity)
		if err != nil {
			return errs.NewDalError("BaseRepository.Change", "change entity", err)
		}

		res, err = br.callbacks.RowScanner(row)
		if err != nil {
			if br.db.GetHelper().IsUniqueViolation(err) {
				return errs.NewDalAlreadyExistsError(br.info.Entity, (*entity).GetID(), err)
			}

			return errs.NewDalError("BaseRepository.Change", "scan after change entity", err)
		}

		return nil
	})
	if err != nil {
		return nil, errs.NewDalError("BaseRepository.Change", "run in transaction", err)
	}

	if br.callbacks.AfterFind != nil {
		return br.callbacks.AfterFind(res)
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	if err := br.prepareDelete(); err != nil {
		return errs.NewNotImplementedError(err)
	}

	err := br.db.GetHelper().ExecStmt(ctx, br.deleteStmt, func(err error) (string, any, error) {
		return br.info.Entity, id, err
	}, id)
	if err != nil {
		return errs.NewDalError("BaseRepository.Delete", "delete row", err)
	}

	return nil
}

func (br *BaseRepository[T, ID]) prepareDelete() error {
	var err error = nil
	br.onceDelete.Do(func() {
		if br.queryBuilders == nil {
			err = errs.NewDalError("BaseRepository.prepareDelete", "query builders not applied", nil)
		}
		if br.queryBuilders.findBuilder == nil {
			err = errs.NewDalError("BaseRepository.prepareDelete", "query find builder not applied", nil)
		}
		sqlFind := br.queryBuilders.findBuilder()
		if strings.TrimSpace(sqlFind) == "" {
			err = errs.NewDalError("BaseRepository.prepareDelete", "query delete empty", nil)
		}

		queryCtx, queryCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer queryCancel()

		br.deleteStmt, err = br.db.GetDB().PrepareContext(queryCtx, br.queryBuilders.GetDelete()())
		if err != nil {
			err = errs.NewDalError("BaseRepository.prepareDelete", "prepare delete stmt", err)
		}
	})

	return nil
}

func (br *BaseRepository[T, ID]) Close() error {
	massErrors := errors.Join(br.findStmt.Close(), br.deleteStmt.Close())
	if massErrors != nil {
		return errs.NewDalError("BaseRepository.Close", "close resources", massErrors)
	}

	return nil
}

func (br *BaseRepository[T, ID]) getInfo() *EntityInfo {
	return br.info
}

func (br *BaseRepository[T, ID]) GetDB() db.DB {
	return br.db
}

func (br *BaseRepository[T, ID]) GetDBHelper() helper.DBHelper {
	return br.db.GetHelper()
}

func (br *BaseRepository[T, ID]) GetQueryBuilders() *BaseQueryBuilders {
	return br.queryBuilders
}

func (br *BaseRepository[T, ID]) GetCallbacks() *BaseRepositoryCallbacks[T, ID] {
	return br.callbacks
}
