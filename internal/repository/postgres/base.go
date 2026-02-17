package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// Дополнительные методы call back
type (
	RowScannerFunc[T domain.Identity[ID], ID any]         func(row *sql.Row) (*T, error)
	RowsScannerFunc[T domain.Identity[ID], ID any]        func(ctx context.Context, rows *sql.Rows, dest *T) error
	AfterGetFunc[T domain.Identity[ID], ID any]           func(*T) (*T, error)
	AfterListYeldFunc[T domain.Identity[ID], ID any]      func(*T) (*T, error)
	NewEntityFactory[T domain.Identity[ID], ID any]       func() *T
	ValidateEntityFunc[T domain.Identity[ID], ID any]     func(*T) error
	BeforeCreateChangeFunc[T domain.Identity[ID], ID any] func(*T) error
)

type BaseRepository[T domain.Identity[ID], ID any] struct {
	db     db.DB
	table  string
	entity string

	getStmt    *sql.Stmt
	deleteStmt *sql.Stmt

	rowScanner  RowScannerFunc[T, ID]
	rowsScanner RowsScannerFunc[T, ID]

	afterGet         AfterGetFunc[T, ID]
	afterListYeld    AfterListYeldFunc[T, ID]
	newEntityFactory NewEntityFactory[T, ID]

	validateCreate ValidateEntityFunc[T, ID]
	validateChange ValidateEntityFunc[T, ID]

	beforeCreate BeforeCreateChangeFunc[T, ID]
	beforeChange BeforeCreateChangeFunc[T, ID]
}

//goland:noinspection GoResourceLeak
func NewBaseRepository[T domain.Identity[ID], ID any](
	db db.DB,
	table,
	entity string,
	rowScanner RowScannerFunc[T, ID],
	rowsScanner RowsScannerFunc[T, ID],
	afterGet AfterGetFunc[T, ID],
	afterListYeld AfterListYeldFunc[T, ID],
	newEntityFactory NewEntityFactory[T, ID],
) (*BaseRepository[T, ID], error) {
	ctxStmt, cancelStmt := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStmt()

	getStmt, err := db.GetDB().PrepareContext(ctxStmt, fmt.Sprintf("select * from %s where id = $1", table))
	if err != nil {
		return nil, errs.NewDalError("NewBaseRepository", "prepare get stmt", err)
	}
	deleteStmt, err := db.GetDB().PrepareContext(ctxStmt, fmt.Sprintf("delete from %s where id = $1", table))
	if err != nil {
		return nil, errs.NewDalError("NewBaseRepository", "prepare delete stmt", err)
	}

	return &BaseRepository[T, ID]{
		db:               db,
		table:            table,
		entity:           entity,
		getStmt:          getStmt,
		deleteStmt:       deleteStmt,
		rowScanner:       rowScanner,
		rowsScanner:      rowsScanner,
		afterGet:         afterGet,
		afterListYeld:    afterListYeld,
		newEntityFactory: newEntityFactory,
	}, nil
}

func (br *BaseRepository[T, ID]) Get(ctx context.Context, id ID) (*T, error) {
	res, err := br.rowScanner(br.getStmt.QueryRowContext(ctx, id))
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return nil, errs.NewDalNotFoundError(br.entity, id, err)
		}

		return nil, errs.NewDalError("BaseRepository.Get", "get row", err)
	}

	if br.afterGet != nil {
		return br.afterGet(res)
	}

	return res, nil
}

func (br *BaseRepository[T, ID]) List(ctx context.Context, limit, offset int) ([]*T, error) {
	return br.InternalList(ctx, fmt.Sprintf("select * from %s order by id asc limit $1 offset $2", br.table), limit, offset)
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

		entity := br.newEntityFactory()

		err = br.rowsScanner(ctx, rows, entity)
		if err != nil {
			return nil, errs.NewDalError("BaseRepository.InternalList", "scan rows", err)
		}

		if br.afterListYeld != nil {
			entity, err = br.afterListYeld(entity)
			if err != nil {
				return nil, errs.NewDalError("BaseRepository.InternalList", "post scan processing", err)
			}
		}
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
	if br.validateCreate != nil {
		if err := br.validateCreate(entity); err != nil {
			return nil, errs.NewDalError("BaseRepository.Create", "validate create", err)
		}
	}

	if br.beforeCreate != nil {
		if err := br.beforeCreate(entity); err != nil {
			return nil, errs.NewDalError("BaseRepository.Create", "before create", err)
		}
	}

	// выполнение
	err := br.db.GetHelper().RunInTx(ctx, br.db.GetDB(), func(ctx context.Context, tx *sql.Tx) error {
		// стейтмент
		stmt, err := tx.PrepareContext(ctx, sqlUserCreate)
		if err != nil {
			return apperrs.NewDalCommonError("UserRepo.Create", "prepare stmt", err)
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(ctx,
			user.ID,
			user.Key.Username,
			user.PasswordHash,
			user.PrivateKey,
			user.PublicKey,
			user.Active,
			user.Deleted,
			user.Person,
			user.EMail,
		)

		if err != nil {
			if urp.db.GetHelper().IsUniqueViolation(err) {
				return apperrs.NewDalAlreadyExistsError("User", user.Key.Username, err)
			}

			return apperrs.NewDalCommonError("UserRepo.Create", "exec stmt", err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrs.NewDalCommonError("UserRepo.Create", "run in transaction", err)
	}

	return urp.afterGet(user)
}

func (br *BaseRepository[T, ID]) Change(ctx context.Context, entity *T) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (br *BaseRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	err := br.db.GetHelper().ExecStmt(ctx, br.deleteStmt, func(err error) (string, any, error) {
		return br.entity, id, err
	}, id)
	if err != nil {
		return errs.NewDalError("BaseRepository.Delete", "delete row", err)
	}

	return nil
}

func (br *BaseRepository[T, ID]) Close() error {
	massErrors := errors.Join(br.getStmt.Close(), br.deleteStmt.Close())
	if massErrors != nil {
		return errs.NewDalError("Close", "close resources", massErrors)
	}

	return nil
}
