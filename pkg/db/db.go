package db

import (
	"context"
	"database/sql"
)

type NotFoundInfo func(error) (string, any, error)

type Executor interface {
	GetQuerier(ctx context.Context) Querier
}

type ErrorDecipher interface {
	IsUniqueViolation(err error) bool
	//IsForeignKeyViolation(err error) bool
	// Можно добавить IsConnectionError, IsDeadlock и т.д.
}

type DB interface {
	Executor
	ErrorDecipher
	GetDriver() string
	GetDB() *sql.DB
	GetDSN() string
	Close() error
}
