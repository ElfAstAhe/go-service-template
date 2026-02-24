package example_service

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/pressly/goose/v3"
)

const (
	sqlCreateTableTest = `
create table if not exists test (
    id varchar(50) not null,
    code varchar(50) not null,
    name varchar(100) null,
    description varchar(512) null,
    created_at datetimetz not null default now(),
    updated_at datetimetz not null default now(),
    constraint test_pk primary key (id),
    constraint test_uk unique (code)
)
`
	sqlDropTableTest = `
drop table if exists test
`
	sqlCreateIndexTestCode = `create index if not exists idx_test_code on test (code asc)`
	sqlDropIndexTestCode   = `drop index if exists idx_test_code`
)

func up0001(ctx context.Context, db *sql.DB) error {
	if err := upCreateTableTest(ctx, db); err != nil {
		return err
	}
	if err := upCreateIndexTestCode(ctx, db); err != nil {
		return err
	}

	return nil
}

func upCreateTableTest(ctx context.Context, db *sql.DB) error {
	if _, err := db.Exec(sqlCreateTableTest); err != nil {
		return errs.NewDBMigrationError("create table test", err)
	}

	return nil
}

func upCreateIndexTestCode(ctx context.Context, db *sql.DB) error {
	if _, err := db.Exec(sqlCreateIndexTestCode); err != nil {
		return errs.NewDBMigrationError("create index idx_test_code", err)
	}

	return nil
}

func down0001(ctx context.Context, db *sql.DB) error {
	if err := downDropIndexTestCode(ctx, db); err != nil {
		return err
	}
	if err := downDropTableTest(ctx, db); err != nil {
		return err
	}

	return nil
}

func downDropIndexTestCode(ctx context.Context, db *sql.DB) error {
	if _, err := db.Exec(sqlDropIndexTestCode); err != nil {
		return errs.NewDBMigrationError("drop index idx_test_code", err)
	}

	return nil
}

func downDropTableTest(ctx context.Context, db *sql.DB) error {
	if _, err := db.Exec(sqlDropTableTest); err != nil {
		return errs.NewDBMigrationError("drop table test", err)
	}

	return nil
}

func init() {
	goose.AddMigrationNoTxContext(up0001, down0001)
}
