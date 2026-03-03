package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/XSAM/otelsql"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/xo/dburl"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

type PgDB struct {
	db   *sql.DB
	conf *config.DBConfig
}

func NewPgDB(conf *config.DBConfig) (*PgDB, error) {
	pg, err := sql.Open("pgx", conf.DSN)
	if err != nil {
		return nil, errs.NewDalError("NewPgDB", "failed to open pgx db", err)
	}

	appDB, err := setupDB(pg, conf)
	if err != nil {
		return nil, errs.NewDalError("NewPgDB", "failed setup db connection", err)
	}

	return appDB, nil
}

func NewPgDBTracing(conf *config.DBConfig) (*PgDB, error) {
	u, err := dburl.Parse(conf.DSN)
	if err != nil {
		return nil, errs.NewDalError("NewPgDBTracing", "parse DSN", err)
	}
	// Вместо sql.Open("postgres", ...) делаем:
	pg, err := otelsql.Open("postgres", conf.DSN,
		otelsql.WithAttributes(
			semconv.DBSystemNamePostgreSQL,
			semconv.DBSystemNameKey.String(u.Path),
		),
		// Включаем трейсинг всех запросов к БД
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			Ping: true, // Трейсить даже проверки связи (healthchecks)
		}),
	)
	if err != nil {
		return nil, errs.NewDalError("NewPgDBTracing", "failed to open otel pgx db", err)
	}

	appDB, err := setupDB(pg, conf)
	if err != nil {
		return nil, errs.NewDalError("NewPgDBTracing", "failed setup db connection", err)
	}

	return appDB, nil
}

func setupDB(pg *sql.DB, conf *config.DBConfig) (*PgDB, error) {
	pg.SetMaxOpenConns(conf.MaxOpenConns)
	pg.SetMaxIdleConns(conf.MaxIdleConns)
	pg.SetConnMaxIdleTime(conf.ConnMaxIdleLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), conf.ConnTimeout)
	defer cancel()

	err := pg.PingContext(ctx)
	if err != nil {
		return nil, errs.NewDalError("setupDB", "ping db connection", err)
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
