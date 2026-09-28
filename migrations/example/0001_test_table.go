package example

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/pressly/goose/v3"
)

func up0001(ctx context.Context, db *sql.DB) error {
	if err := upCreateTableTest(ctx, db); err != nil {
		return err
	}

	return upCreateIndexTestCode(ctx, db)
}

func upCreateTableTest(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, sqlCreateTableTest); err != nil {
		return errs.NewDBMigrationError("create table test", err)
	}

	return nil
}

func upCreateIndexTestCode(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, sqlCreateIndexTestCode); err != nil {
		return errs.NewDBMigrationError("create index idx_test_code", err)
	}

	return nil
}

func down0001(ctx context.Context, db *sql.DB) error {
	if err := downDropIndexTestCode(ctx, db); err != nil {
		return err
	}
	return downDropTableTest(ctx, db)
}

func downDropIndexTestCode(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, sqlDropIndexTestCode); err != nil {
		return errs.NewDBMigrationError("drop index idx_test_code", err)
	}

	return nil
}

func downDropTableTest(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, sqlDropTableTest); err != nil {
		return errs.NewDBMigrationError("drop table test", err)
	}

	return nil
}

func init() {
	goose.AddMigrationNoTxContext(up0001, down0001)
}
