package example

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/pressly/goose/v3"
)

func up0002(ctx context.Context, db *sql.DB) error {
	if err := createTableTest2(ctx, db); err != nil {
		return err
	}

	return createIndexTest2Code(ctx, db)
}

func createTableTest2(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sqlCreateTableTest2)
	if err != nil {
		return errs.NewDBMigrationError("create table test2 failed", err)
	}

	return nil
}

func createIndexTest2Code(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sqlCreateIndexTest2Code)
	if err != nil {
		return errs.NewDBMigrationError("create index test2 failed", err)
	}

	return nil
}

func down0002(ctx context.Context, db *sql.DB) error {
	if err := dropIndexTest2Code(ctx, db); err != nil {
		return err
	}

	return dropTableTest2(ctx, db)
}

func dropTableTest2(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sqlDropTableTest2)
	if err != nil {
		return errs.NewDBMigrationError("drop table test2 failed", err)
	}

	return nil
}

func dropIndexTest2Code(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sqlDropIndexTest2Code)
	if err != nil {
		return errs.NewDBMigrationError("drop index test2 failed", err)
	}

	return nil
}

func init() {
	goose.AddMigrationNoTxContext(up0002, down0002)
}
