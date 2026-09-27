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

// DBMigrator реализует интерфейс migration.Migrator, используя в качестве движка утилиту Goose v3.
//
// Координирует безопасное, транзакционное применение SQL и Go-скриптов изменения схемы данных (DDL),
// обеспечивая эволюцию структуры СУБД (PostgreSQL) в cloud-native средах.
type DBMigrator struct {
	db  db.DB         // Ссылка на глобальный пул и конфигурацию соединения с БД
	log logger.Logger // Изолированный системный логгер компонента
}

// Гарантируем соответствие контракту Migrator на этапе компиляции
var _ migration.Migrator = (*DBMigrator)(nil)

// NewDBMigrator — фабричный конструктор мигратора Goose.
func NewDBMigrator(db db.DB, logger logger.Logger) (*DBMigrator, error) {
	return &DBMigrator{
		db:  db,
		log: logger.GetLogger("DB migrator"),
	}, nil
}

// Initialize конфигурирует глобальные параметры рантайма Goose.
// Настраивает SQL-диалект драйвера, переопределяет системную таблицу истории версий миграций
// и инжектирует кастомный структурированный логгер фреймворка.
func (g *DBMigrator) Initialize() error {
	if err := goose.SetDialect(g.db.GetDriver()); err != nil {
		return errs.NewDBMigrationError("error select dialect", err)
	}
	goose.SetTableName("goose_version_history")
	goose.SetLogger(logger.NewGooseLogger(g.log))

	return nil
}

// Up атомарно применяет все новые файлы миграций, переводя схему БД на актуальное состояние.
// Защищен от фатальных паник рантайма (Panic Recovery паттерн).
func (g *DBMigrator) Up(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Проверяем тип объекта паники и безопасно приводим его к интерфейсу error
			recoveryErr, ok := r.(error)
			if !ok {
				recoveryErr = errs.NewConfigError(fmt.Sprintf("panic [%v] recovery", r), nil)
			}
			err = errs.NewDBMigrationError("migrate up panic", recoveryErr)
		}
	}()

	// Опция WithAllowMissing() разрешает накат старых пропущенных миграций (Out-of-order migrations)
	if err := goose.UpContext(ctx, g.db.GetDB(), ".", goose.WithAllowMissing()); err != nil {
		return errs.NewDBMigrationError("error migrate up", err)
	}

	return nil
}

// Down осуществляет последовательный откат схемы базы данных на одну версию назад.
func (g *DBMigrator) Down(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryErr, ok := r.(error)
			if !ok {
				recoveryErr = errs.NewConfigError(fmt.Sprintf("panic [%v] recovery", r), nil)
			}
			// ИСПРАВЛЕНО: Текст ошибки изменен на маркер "migrate down panic"
			err = errs.NewDBMigrationError("migrate down panic", recoveryErr)
		}
	}()

	if err := goose.DownContext(ctx, g.db.GetDB(), ".", goose.WithAllowMissing()); err != nil {
		return errs.NewDBMigrationError("error migrate down", err)
	}

	return nil
}
