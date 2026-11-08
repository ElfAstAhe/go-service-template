package postgres

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CreateTableDataAuditOutbox инициализирует создание унифицированной локальной таблицы data_audit_outbox.
// Вызывается на фазе холодного старта или при выполнении миграций схемы данных микросервиса.
func CreateTableDataAuditOutbox(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, sqlCreateTableDataAuditOutbox); err != nil {
		return errs.NewDBMigrationError("failed to create table data_audit_outbox", err)
	}

	return nil
}

// CreateIndexDataAuditOutboxPending конструирует специализированный частичный индекс для оптимизации сквозной выборки логов.
// Гарантирует lock-free производительность и субмиллисекундное выполнение оператора SKIP LOCKED.
func CreateIndexDataAuditOutboxPending(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, sqlCreateIndexDataAuditOutboxPending); err != nil {
		return errs.NewDBMigrationError("failed to create index idx_data_audit_outbox_pending", err)
	}

	return nil
}

// DropTableDataAuditOutbox выполняет деструктивное удаление таблицы data_audit_outbox из схемы данных СУБД.
func DropTableDataAuditOutbox(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, sqlDropTableDataAuditOutbox); err != nil {
		return errs.NewDBMigrationError("failed to drop table data_audit_outbox", err)
	}

	return nil
}

// DropIndexDataAuditOutboxPending выполняет удаление частичного индекса оптимизации выборки из схемы данных СУБД.
func DropIndexDataAuditOutboxPending(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, sqlDropIndexDataAuditOutboxPending); err != nil {
		return errs.NewDBMigrationError("failed to drop index idx_data_audit_outbox_pending", err)
	}

	return nil
}
