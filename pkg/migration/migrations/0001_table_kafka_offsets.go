package migrations

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

func Up0001Kafka(ctx context.Context, db *sql.DB) error {
	if err := createTableKafkaOffsets(ctx, db); err != nil {
		return err
	}
	if err := createIndexKafkaOffsetsLookup(ctx, db); err != nil {
		return err
	}

	return nil
}

func createTableKafkaOffsets(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sqlCreateTableKafkaOffsets)
	if err != nil {
		return errs.NewDBMigrationError("create table kafka offsets", err)
	}

	return nil
}

func createIndexKafkaOffsetsLookup(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sqlCreateIndexKafkaOffsets)
	if err != nil {
		return errs.NewDBMigrationError("create table kafka offsets lookup", err)
	}

	return nil
}

func Down0001Kafka(ctx context.Context, db *sql.DB) error {
	if err := dropIndexKafkaOffsetsLookup(ctx, db); err != nil {
		return err
	}
	if err := dropTableKafkaOffsets(ctx, db); err != nil {
		return err
	}

	return nil
}

func dropTableKafkaOffsets(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sqlDropTableKafkaOffsets)
	if err != nil {
		return errs.NewDBMigrationError("drop table kafka offsets", err)
	}

	return nil
}

func dropIndexKafkaOffsetsLookup(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sqlDropIndexKafkaOffsets)
	if err != nil {
		return errs.NewDBMigrationError("drop table kafka offsets lookup", err)
	}

	return nil
}
