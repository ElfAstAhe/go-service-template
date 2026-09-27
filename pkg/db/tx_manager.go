package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// Приватный тип ключа контекста полностью исключает коллизии (Context Collision)
// при пробросе транзакции сквозь слои приложения.
type txKeyType struct{}

var txKey txKeyType = txKeyType{}

// TxManager реализует интерфейс TransactionManager, обеспечивая декларативное,
// потокобезопасное и транзакционное выполнение доменных операций UseCase-слоя.
type TxManager struct {
	db DB // Ссылка на глобальный интерфейс подсистемы СУБД
}

// Гарантируем соответствие контракту TransactionManager на этапе компиляции
var _ TransactionManager = (*TxManager)(nil)

// NewTxManager — фабричный конструктор менеджера транзакций.
func NewTxManager(db DB) *TxManager {
	return &TxManager{
		db: db,
	}
}

// WithinTransaction оборачивает выполнение прикладной функции fn в ACID-транзакцию СУБД.
// Автоматически управляет вызовами Begin, Commit и Rollback на основе результатов работы функции или паник рантайма.
func (tm *TxManager) WithinTransaction(ctx context.Context, opts *TransactionOptions, fn func(ctx context.Context) error) (err error) {
	// 🛡️ Защитный барьер (Propagation): если в контексте уже есть открытая транзакция,
	// повторно Begin не вызываем, а прозрачно прокидываем управление дальше.
	if tx := GetTx(ctx); tx != nil {
		return fn(ctx)
	}

	var sqlOpts *sql.TxOptions
	if opts != nil {
		sqlOpts = &sql.TxOptions{
			Isolation: mapIsolationLevelSQLIsolation(opts.Isolation), // Маппим абстрактный уровень на стандарт Go
			ReadOnly:  opts.ReadOnly,
		}
	}

	// Инициируем физическое открытие транзакции в пуле соединений СУБД
	tx, err := tm.db.GetDB().BeginTx(ctx, sqlOpts)
	if err != nil {
		return errs.NewDalError("TxManager.WithinTransaction", "error begin transaction", err)
	}

	// Автоматический оркестратор жизненного цикла транзакции
	defer func() {
		if r := recover(); r != nil {
			// Сценарий 1: Бизнес-код UseCase-а упал с паникой. Немедленно откатываем транзакцию СУБД.
			_ = tx.Rollback()

			var recoveryErr error
			if e, ok := r.(error); ok {
				recoveryErr = e
			} else {
				recoveryErr = fmt.Errorf("recovery [%v]", r)
			}

			// ИСПРАВЛЕНО: Текст операции теперь строго указывает на текущий метод TxManager.WithinTransaction
			err = errs.NewDalError("TxManager.WithinTransaction", "panic recovery", recoveryErr)
		} else if err != nil {
			// Сценарий 2: Функция завершилась штатно, но вернула ошибку бизнеса/валидации. Делаем Rollback.
			_ = tx.Rollback()
		} else {
			// Сценарий 3: Всё прошло идеально. Фиксируем транзакцию (Commit).
			err = tx.Commit()
			if err != nil {
				err = errs.NewDalError("TxManager.WithinTransaction", "commit", err)
			}
		}
	}()

	// Упаковываем указатель на открытую транзакцию в контекст для нижележащих репозиториев
	txCtx := context.WithValue(ctx, txKey, tx)

	// Передаем управление прикладной логике UseCase
	err = fn(txCtx)

	return err
}

// GetTx извлекает нативный объект *sql.Tx из контекста. Возвращает nil, если вызов происходит вне транзакции.
func GetTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}

	return nil
}

// mapIsolationLevelSQLIsolation транслирует агностичные уровни изоляции фреймворка в системные константы database/sql.
func mapIsolationLevelSQLIsolation(level IsolationLevel) sql.IsolationLevel {
	switch level {
	case LevelReadCommitted:
		return sql.LevelReadCommitted
	case LevelRepeatableRead:
		return sql.LevelRepeatableRead
	case LevelSerializable:
		return sql.LevelSerializable
	default:
		return sql.LevelDefault
	}
}
