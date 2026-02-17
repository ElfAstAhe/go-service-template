package helper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresDBHelper struct{}

func NewPostgresDBHelper() *PostgresDBHelper {
	return &PostgresDBHelper{}
}

// IsUniqueViolation проверяет, является ли ошибка нарушением уникальности
func (pdh *PostgresDBHelper) IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // Код ошибки unique_violation в PostgreSQL
	}

	return false
}

// RunInTx инкапсулирует всю грязную работу с транзакциями.
// Принимает контекст, инстанс БД и функцию, которую нужно выполнить в транзакции.
func (pgh *PostgresDBHelper) RunInTx(ctx context.Context, db *sql.DB, fn func(ctx context.Context, tx *sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errs.NewDalError("PosgtesDBHelper.RunInTx", "begin transaction", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback() // Откатываем в любом случае

			// Превращаем панику в читаемую ошибку для логов
			var recoveryErr error
			if e, ok := r.(error); ok {
				recoveryErr = e
			} else {
				recoveryErr = fmt.Errorf("recovery [%v]", r)
			}

			err = errs.NewDalError("PosgtesDBHelper.RunInTx", "panic recovery", recoveryErr)
		} else if err != nil {
			_ = tx.Rollback() // Откат при ошибке бизнеса/БД
		} else {
			err = tx.Commit() // Фиксация
			if err != nil {
				err = errs.NewDalError("PosgtesDBHelper.RunInTx", "commit", err)
			}
		}
	}()

	err = fn(ctx, tx)

	return err
}

func (pgh *PostgresDBHelper) ExecStmt(ctx context.Context, stmt *sql.Stmt, notFoundInfo NotFoundInfo, params ...any) error {
	// выполняем
	res, err := stmt.ExecContext(ctx, params...)
	if err != nil {
		return errs.NewDalError("PosgtesDBHelper.ExecStmt", "exec ctx", err)
	}
	// проверяем
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errs.NewDalError("PosgtesDBHelper.ExecStmt", "rows affected", err)
	}
	if !(rowsAffected > 0) {
		return errs.NewDalNotFoundError(notFoundInfo(err))
	}

	return nil
}
