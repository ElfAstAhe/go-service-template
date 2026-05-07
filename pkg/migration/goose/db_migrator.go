package goose

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/migration"
	"github.com/pressly/goose/v3"
)

// DBMigrator is implementation of DBMigrator interface
type DBMigrator struct {
	db  db.DB
	log logger.Logger
}

var _ migration.Migrator = (*DBMigrator)(nil)

func NewDBMigrator(db db.DB, logger logger.Logger) (*DBMigrator, error) {
	return &DBMigrator{
		db:  db,
		log: logger.GetLogger("DB migrator"),
	}, nil
}

func (g *DBMigrator) Initialize() error {
	if err := goose.SetDialect(g.db.GetDriver()); err != nil {
		return errs.NewDBMigrationError("error select dialect", err)
	}
	goose.SetTableName("goose_version_history")
	goose.SetLogger(logger.NewGooseLogger(g.log))

	return nil
}

func (g *DBMigrator) Up(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Проверяем, является ли r ошибкой
			recoveryErr, ok := r.(error)
			if !ok {
				// Если это строка или что-то другое, приводим к виду error вручную
				recoveryErr = errs.NewConfigError(fmt.Sprintf("panic [%v] recovery", r), nil)
			}
			err = errs.NewDBMigrationError("migrate up panic", recoveryErr)
		}
	}()
	if err := goose.UpContext(ctx, g.db.GetDB(), ".", goose.WithAllowMissing()); err != nil {
		return errs.NewDBMigrationError("error migrate up", err)
	}

	return nil
}

func (g *DBMigrator) Down(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Проверяем, является ли r ошибкой
			recoveryErr, ok := r.(error)
			if !ok {
				// Если это строка или что-то другое, приводим к виду error вручную
				recoveryErr = errs.NewConfigError(fmt.Sprintf("panic [%v] recovery", r), nil)
			}
			err = errs.NewDBMigrationError("migrate up panic", recoveryErr)
		}
	}()
	if err := goose.DownContext(ctx, g.db.GetDB(), ".", goose.WithAllowMissing()); err != nil {
		return errs.NewDBMigrationError("error migrate down", err)
	}

	return nil
}
