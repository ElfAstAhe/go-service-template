package helper

import (
	"context"
	"database/sql"
)

type NotFoundInfo func(error) (string, string, error)

type DBHelper interface {
	RunInTx(ctx context.Context, db *sql.DB, fn func(ctx context.Context, tx *sql.Tx) error) error
	ExecStmt(ctx context.Context, stmt *sql.Stmt, notFoundInfo NotFoundInfo, params ...any) error

	IsUniqueViolation(err error) bool
}
