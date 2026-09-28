package db

import (
	"context"
	"database/sql"
)

// Executor определяет контракт для динамического извлечения исполнителя SQL-запросов.
// Обеспечивает прозрачное переключение между голыми соединениями и ACID-транзакциями на уровне контекста.
type Executor interface {
	// GetQuerier возвращает активный объект Querier (либо транзакцию *sql.Tx, либо пул сокетов *sql.DB),
	// привязанный к текущему контексту выполнения запроса context.Context.
	GetQuerier(ctx context.Context) Querier
}

// ErrorDecipher описывает контракт переводчика (дешифратора) низкоуровневых системных ошибок СУБД.
// Абстрагирует слой данных (DAL) от вендоро-специфичных кодов ошибок конкретных баз данных (PostgreSQL/MySQL).
type ErrorDecipher interface {
	// IsUniqueViolation проверяет, вызван ли сбой нарушением уникального индекса (Unique Constraint Violation).
	IsUniqueViolation(err error) bool
	// IsForeignKeyViolation(err error) bool
	// Можно добавить IsConnectionError, IsDeadlock и т.д.
}

// DB объединяет под своей эгидой управление пулом соединений, механизмы дешифрации ошибок СУБД
// и контроль жизненного цикла физических сокетов базы данных.
type DB interface {
	Executor
	ErrorDecipher

	// GetDriver возвращает строковый маркер используемого SQL-драйвера (например, "postgres").
	GetDriver() string

	// GetDB возвращает ссылку на нативный низкоуровневый пул соединений стандартной библиотеки Go.
	GetDB() *sql.DB

	// GetDSN возвращает строку подключения к источнику данных (Data Source Name) с маскированными паролями.
	GetDSN() string

	// Ping проверяет физическую доступность СУБД и целостность сетевого канала вызовом тестового пинга.
	Ping(ctx context.Context) error

	// Close осуществляет мягкую остановку пула, принудительно разрывая все активные сетевые сокеты с базой.
	Close() error
}
