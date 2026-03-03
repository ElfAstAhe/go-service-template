package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type txKeyType struct{}

var txKey txKeyType = txKeyType{}

type TxManager struct {
	db DB
}

func NewTxManager(db DB) *TxManager {
	return &TxManager{
		db: db,
	}
}

func (tm *TxManager) WithinTransaction(ctx context.Context, opts *TransactionOptions, fn func(ctx context.Context) error) (err error) {
	if tx := GetTx(ctx); tx != nil {
		return fn(ctx)
	}

	var sqlOpts *sql.TxOptions
	if opts != nil {
		sqlOpts = &sql.TxOptions{
			Isolation: mapIsolationLevelSqlIsolation(opts.Isolation),
			ReadOnly:  opts.ReadOnly,
		}
	}

	tx, err := tm.db.GetDB().BeginTx(ctx, sqlOpts)
	if err != nil {
		return errs.NewDalError("TxManager.WithTransaction", "error begin transaction", err)
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

	txCtx := context.WithValue(ctx, txKey, tx)

	err = fn(txCtx)

	return err
}

func GetTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}

	return nil
}

func mapIsolationLevelSqlIsolation(level IsolationLevel) sql.IsolationLevel {
	switch level {
	case LevelReadCommitted:
		return sql.LevelReadCommitted
	case LevelRepeatableRead:
		return sql.LevelRepeatableRead
	case LevelSerializable:
		return sql.LevelSerializable
	default:
		return sql.LevelDefault
	}
}
