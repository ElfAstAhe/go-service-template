package migrations

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/pressly/goose/v3"
)

// GooseDBMigrator is implementation of DBMigrator interface
type GooseDBMigrator struct {
	db  db.DB
	ctx context.Context
	log logger.Logger
}

func NewGooseDBMigrator(ctx context.Context, db db.DB, logger logger.Logger) (*GooseDBMigrator, error) {
	return &GooseDBMigrator{
		db:  db,
		ctx: ctx,
		log: logger.GetLogger("DB migrator"),
	}, nil
}

// DBMigrator

func (g *GooseDBMigrator) Initialize() error {
	if err := goose.SetDialect(g.db.GetDriver()); err != nil {
		return errs.NewDBMigrationError("error select dialect", err)
	}
	goose.SetTableName("goose_version_history")
	goose.SetLogger(logger.NewGooseLogger(g.log))

	return nil
}

func (g *GooseDBMigrator) Up() error {
	if err := goose.UpContext(g.ctx, g.db.GetDB(), ".", goose.WithAllowMissing()); err != nil {
		return errs.NewDBMigrationError("error migrate up", err)
	}

	return nil
}

func (g *GooseDBMigrator) Down() error {
	if err := goose.DownContext(g.ctx, g.db.GetDB(), ".", goose.WithAllowMissing()); err != nil {
		return errs.NewDBMigrationError("error migrate down", err)
	}

	return nil
}

// ==============
