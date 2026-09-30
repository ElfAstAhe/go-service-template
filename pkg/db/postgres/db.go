package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/jackc/pgx/v5/pgconn"
)

// DB implements db.DB, db.Executor, and db.ErrorDecipher interfaces,
// serving as a production-grade relational database client for PostgreSQL.
// Wraps standard sql.DB handle fortified with pgx driver configurations under the hood.
type DB struct {
	db   *sql.DB
	conf *config.DBConfig
}

// Compile-time interface compliance verifications
var _ db.DB = (*DB)(nil)
var _ db.Executor = (*DB)(nil)
var _ db.ErrorDecipher = (*DB)(nil)

// New acts as a factory constructor establishing a connection pool to PostgreSQL via pgx driver.
// Performs early validation and structures standard connection errors into internal DalError taxonomies.
func New(conf *config.DBConfig) (*DB, error) {
	pg, err := sql.Open("pgx", conf.DSN)
	if err != nil {
		return nil, errs.NewDalError("NewDB", "failed to open pgx db", err)
	}

	appDB, err := Setup(pg, conf)
	if err != nil {
		return nil, errs.NewDalError("NewPgDB", "failed setup db connection", err)
	}

	return appDB, nil
}

// Setup initializes connection pool limits including max open/idle counts and lifetime bounds.
func Setup(pg *sql.DB, conf *config.DBConfig) (*DB, error) {
	pg.SetMaxIdleConns(conf.MaxIdleConns)
	pg.SetMaxOpenConns(conf.MaxOpenConns)
	pg.SetConnMaxIdleTime(conf.ConnMaxIdleLifetime)

	return &DB{
		db:   pg,
		conf: conf,
	}, nil
}

// GetDriver extracts the configuration string identifying the underlying sql driver name.
func (pgd *DB) GetDriver() string {
	return pgd.conf.Driver
}

// GetDB yields a direct reference to the managed underlying *sql.DB connection pool handler.
func (pgd *DB) GetDB() *sql.DB {
	return pgd.db
}

// GetDSN extracts the active Data Source Name string used to authenticate connection sockets.
func (pgd *DB) GetDSN() string {
	return pgd.conf.DSN
}

// Close gracefully terminates all active network connections present in the database pool.
func (pgd *DB) Close() error {
	return pgd.db.Close()
}

// GetQuerier inspects the execution context to determine if a context-bound transaction is active.
// Returns an active transaction handle if present; otherwise, falls back to the standard non-transactional db pool.
func (pgd *DB) GetQuerier(ctx context.Context) db.Querier {
	if tx := db.GetTx(ctx); tx != nil {
		return tx
	}

	return pgd.db
}

// IsUniqueViolation unwraps the generic error state to evaluate if it corresponds to PostgreSQL code 23505.
func (pgd *DB) IsUniqueViolation(err error) bool {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code == "23505" // Code 23505 identifies unique_violation state in PostgreSQL
	}

	return false
}

// Ping executes a validation round-trip payload to confirm the server is reachable and active.
func (pgd *DB) Ping(ctx context.Context) error {
	err := pgd.db.PingContext(ctx)
	if err != nil {
		return errs.NewDalError("Ping", "ping db connection", err)
	}

	return nil
}
