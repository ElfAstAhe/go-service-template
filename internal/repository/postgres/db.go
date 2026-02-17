package postgres

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/helper"
)

type PgDB struct {
	db     *sql.DB
	conf   *config.DBConfig
	helper helper.DBHelper
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
		db:     pg,
		conf:   conf,
		helper: helper.NewPostgresDBHelper(),
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

func (pgdb *PgDB) GetHelper() helper.DBHelper {
	return pgdb.helper
}

func (pgdb *PgDB) Close() error {
	return pgdb.db.Close()
}
