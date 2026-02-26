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

type PgDB struct {
	db   *sql.DB
	conf *config.DBConfig
}

func NewPgDB(conf *config.DBConfig) (*PgDB, error) {
	pg, err := sql.Open("pgx", conf.DSN)
	if err != nil {
		return nil, err
	}

	pg.SetMaxOpenConns(conf.MaxOpenConns)
	pg.SetMaxIdleConns(conf.MaxIdleConns)
	pg.SetConnMaxIdleTime(conf.ConnMaxIdleLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), conf.ConnTimeout)
	defer cancel()

	err = pg.PingContext(ctx)
	if err != nil {
		return nil, errs.NewDalError("NewPgDB", "ping db connection", err)
	}

	return &PgDB{
		db:   pg,
		conf: conf,
	}, nil
}

func (pgdb *PgDB) GetDriver() string {
	return pgdb.conf.Driver
}

func (pgdb *PgDB) GetDB() *sql.DB {
	return pgdb.db
}

func (pgdb *PgDB) GetDSN() string {
	return pgdb.conf.DSN
}

func (pgdb *PgDB) Close() error {
	return pgdb.db.Close()
}

func (pgdb *PgDB) GetQuerier(ctx context.Context) db.Querier {
	if tx := db.GetTx(ctx); tx != nil {
		return tx
	}

	return pgdb.db
}

func (pgdb *PgDB) IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // Код ошибки unique_violation в PostgreSQL
	}

	return false
}
