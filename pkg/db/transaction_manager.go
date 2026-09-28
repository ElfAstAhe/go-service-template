package db

import (
	"context"
)

// IsolationLevel определяет строго типизированное абстрактное перечисление уровней изоляции транзакций.
// Полностью изолирует прикладной доменный слой от констант стандартного пакета database/sql.
type IsolationLevel int

// Набор констант поддерживаемых стандартов изоляции транзакций (спецификация ANSI SQL).
const (
	LevelDefault        IsolationLevel = iota // Дефолтный уровень изоляции, настроенный на стороне СУБД
	LevelReadCommitted                        // Защита от грязного чтения (Dirty Reads)
	LevelRepeatableRead                       // Защита от грязного и неповторяющегося чтения (Non-Repeatable Reads)
	LevelSerializable                         // Максимальный уровень: полная сериализуемость, защита от фантомов (Phantom Reads)
)

// TransactionOptions инкапсулирует конфигурационные параметры рантайма транзакции.
type TransactionOptions struct {
	Isolation IsolationLevel // Выбранный уровень изоляции данных
	ReadOnly  bool           // Флаг оптимизации транзакции строго на чтение (Read-Only транзакция)
}

// TransactionManager описывает центральный интерфейс управления распределенными транзакциями (Unit of Work).
//
// Абстрагирует UseCase-слой приложения от низкоуровневых вызовов Commit и Rollback.
// Гарантирует атомарное выполнение цепочки операций репозиториев в рамках единой ACID-транзакции СУБД.
type TransactionManager interface {
	// WithinTransaction оборачивает выполнение переданной функции-замыкания fn в ACID-транзакцию СУБД.
	// Если замыкание возвращает ошибку (error != nil) — менеджер автоматически инициирует Rollback.
	// При успешном завершении замыкания менеджер производит финальный Commit изменений.
	WithinTransaction(ctx context.Context, opts *TransactionOptions, fn func(ctx context.Context) error) error
}
