package db

import (
	"context"
)

// IsolationLevel - собственные уровни изоляции, не привязанные к sql
type IsolationLevel int

// Набор констант поддерживаемых уровней изоляции
const (
	LevelDefault IsolationLevel = iota
	LevelReadCommitted
	LevelRepeatableRead
	LevelSerializable
)

// TransactionOptions - опции выполнения в транзакции
type TransactionOptions struct {
	Isolation IsolationLevel
	ReadOnly  bool
}

// TransactionManager - интерфейс, необходим для абстрагирования от реализации
type TransactionManager interface {
	// WithinTransaction выполнение какой-либо операции в рамках транзакции
	WithinTransaction(ctx context.Context, opts *TransactionOptions, fn func(ctx context.Context) error) error
}
