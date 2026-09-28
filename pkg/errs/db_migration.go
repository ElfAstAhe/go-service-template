package errs

import (
	"fmt"
	"runtime"
)

// DBMigrationError реализует нативный интерфейс error, специфицируя системные ошибки,
// возникающие в процессе применения или отката схем баз данных (Database Migration Error).
//
// Ключевой особенностью является автоматический сбор метаданных времени выполнения (runtime):
// фиксирует точное имя файла и строку кода, где была инициализирована ошибка, для ускорения дебага миграций.
type DBMigrationError struct {
	msg  string // Человекочитаемое описание сути сбоя миграции
	err  error  // Ссылка на исходную корневую ошибку (например, синтаксическая ошибка в SQL от драйвера СУБД)
	file string // Путь к файлу кода, инициировавшему ошибку
	line int    // Номер строки кода внутри файла
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*DBMigrationError)(nil)

// NewDBMigrationError — фабричный конструктор ошибки миграционного слоя.
// Автоматически вычисляет метаданные вызывающего потока через рефлексию рантайма (runtime.Caller).
func NewDBMigrationError(msg string, err error) *DBMigrationError {
	dm := &DBMigrationError{
		msg: msg,
		err: err,
	}

	// runtime.Caller(1) извлекает данные о том, какой метод инициировал вызов NewDBMigrationError (шаг назад)
	_, file, line, ok := runtime.Caller(1)
	if ok {
		dm.file = file
		dm.line = line
	}

	return dm
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Интегрирует кликабельные маркеры расположения файла [file.go:line] для бесшовной навигации в IDE.
func (dm *DBMigrationError) Error() string {
	stack := ""
	if dm.file != "" {
		// Формат [file.go:123] является общепринятым стандартом для автоматической генерации ссылок в консоли
		stack = fmt.Sprintf("[%s:%d] ", dm.file, dm.line)
	}

	msg := "DML: migration error"
	if stack != "" {
		msg = fmt.Sprintf("%s at %s", msg, stack)
	}

	if dm.msg != "" {
		msg = fmt.Sprintf("%s, message: %s", msg, dm.msg)
	}

	if dm.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов мигратора (например, Goose)
		msg = fmt.Sprintf("%s: %v", msg, dm.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку мигратора или SQL-драйвера, инициировавшую системный сбой.
// Требуется для корректной работы систем автоматического проброса и десериализации ошибок (errors.Is / errors.As).
func (dm *DBMigrationError) Unwrap() error {
	return dm.err
}
