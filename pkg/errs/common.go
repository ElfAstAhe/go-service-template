package errs

import (
	"fmt"
	"runtime"
)

// CommonError реализует нативный интерфейс error, специфицируя общие системные и утилитарные
// ошибки фреймворка (Common Utility Error), не привязанные жестко к DAL или BLL слоям.
//
// Ключевой особенностью является автоматический lock-free сбор метаданных времени выполнения (runtime):
// фиксирует точное имя файла и строку кода, где была инициализирована ошибка, для ускорения дебага.
type CommonError struct {
	msg  string // Человекочитаемое описание сути технического сбоя
	err  error  // Ссылка на исходную корневую ошибку (если она есть в цепочке вызовов)
	file string // Абсолютный или относительный путь к файлу кода, инициировавшему ошибку
	line int    // Номер строки кода внутри файла
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*CommonError)(nil)

// NewCommonError — фабричный конструктор общей системной ошибки фреймворка.
// Автоматически вычисляет метаданные вызывающего потока через рефлексию рантайма (runtime.Caller).
func NewCommonError(msg string, err error) *CommonError {
	e := &CommonError{
		msg: msg,
		err: err,
	}

	// runtime.Caller(1) извлекает данные о том, какой метод инициировал вызов NewCommonError (шаг назад)
	_, file, line, ok := runtime.Caller(1)
	if ok {
		e.file = file
		e.line = line
	}

	return e
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Интегрирует кликабельные маркеры расположения файла [file.go:line] для бесшовной навигации в IDE.
func (ce *CommonError) Error() string {
	stack := ""
	if ce.file != "" {
		// Формат [file.go:123] является общепринятым стандартом для автоматической генерации ссылок в консоли
		stack = fmt.Sprintf("[%s:%d] ", ce.file, ce.line)
	}

	msg := fmt.Sprintf("CMN: error at %s", stack)
	if ce.msg != "" {
		msg = fmt.Sprintf("%s %s", msg, ce.msg)
	}

	if ce.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, ce.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую системный сбой.
// Обеспечивает корректную работу систем автоматического проброса и десериализации ошибок (errors.Is / errors.As).
func (ce *CommonError) Unwrap() error {
	return ce.err
}
