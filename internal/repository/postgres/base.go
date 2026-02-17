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

type ScannerFunc[T domain.Identity[ID], ID any] func(row *sql.Row) (*T, error)

type ListerFunc[T domain.Identity[ID], ID any] func(ctx context.Context, rows *sql.Rows) ([]*T, error)

type AfterGetFunc[T domain.Identity[ID], ID any] func(*T) (*T, error)

type BaseRepository[T domain.Identity[ID], ID any] struct {
	db     db.DB
	table  string
	entity string

	getStmt    *sql.Stmt
	deleteStmt *sql.Stmt

	scanner  ScannerFunc[T, ID]
	afterGet AfterGetFunc[T, ID]
	lister   ListerFunc[T, ID]
}

//goland:noinspection GoResourceLeak
func NewBaseRepository[T domain.Identity[ID], ID any](
	db db.DB,
	table,
	entity string,
	scanner ScannerFunc[T, ID],
	afterGet AfterGetFunc[T, ID],
	lister ListerFunc[T, ID],
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
		db:         db,
		table:      table,
		entity:     entity,
		getStmt:    getStmt,
		deleteStmt: deleteStmt,
		scanner:    scanner,
		afterGet:   afterGet,
		lister:     lister,
	}, nil
}

func (br *BaseRepository[T, ID]) Get(ctx context.Context, id ID) (*T, error) {
	res, err := br.scanner(br.getStmt.QueryRowContext(ctx, id))
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
	//TODO implement me
	panic("implement me")
}

func (br *BaseRepository[T, ID]) Create(ctx context.Context, entity *T) (*T, error) {
	//TODO implement me
	panic("implement me")
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
