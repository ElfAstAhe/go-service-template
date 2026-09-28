package db

import (
	"context"
	"database/sql"
)

// Querier определяет унифицированный, агностичный интерфейс для исполнения SQL-запросов (Unified Database Interface).
//
// Выступает центральным мостом абстракции, которому нативно удовлетворяют объекты пула (*sql.DB) и транзакций (*sql.Tx).
// Позволяет репозиториям и инфраструктурным хелперам выполнять CRUD-операции без жесткой привязки
// к контексту управления ACID-транзакциями, изолируя логику выполнения от рутины управления соединениями.
type Querier interface {
	// ExecContext выполняет SQL-запрос мутации (INSERT, UPDATE, DELETE), не возвращающий строки выборки.
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	// QueryContext выполняет SQL-запрос чтения (SELECT LIST), возвращающий потоковый набор строк *sql.Rows.
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// QueryRowContext выполняет SQL-запрос точечного чтения (SELECT BY ID), возвращая ровно одну строку *sql.Row.
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

	// PrepareContext компилирует и кэширует шаблон SQL-запроса в СУБД, создавая подготовленное выражение *sql.Stmt.
	// Применяется для оптимизации производительности при циклическом выполнении однотипных команд.
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}
