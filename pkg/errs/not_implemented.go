package errs

import (
	"fmt"
)

// NotImplementedError реализует нативный интерфейс error, специфицируя ошибку обращения
// к еще не реализованному функционалу, заглушкам методов или отсутствующим плагинам фреймворка.
//
// Выступает в роли контролируемого барьера (Guard), предотвращающего паники рантайма
// при вызове пустых конфигураций или методов жизненного цикла.
type NotImplementedError struct {
	err error // Ссылка на нижележащую техническую ошибку, раскрывшую отсутствие реализации
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*NotImplementedError)(nil)

// NewNotImplementedError — фабричный конструктор ошибки отсутствия реализации.
func NewNotImplementedError(err error) *NotImplementedError {
	return &NotImplementedError{err: err}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Выводит общую константу "CMN: not implemented" и аккуратно пристыковывает стек вложенных ошибок.
func (ni *NotImplementedError) Error() string {
	msg := "CMN: not implemented"
	if ni.err != nil {
		// Обогащаем текстовый вывод описанием контекста незавершенной реализации
		msg = fmt.Sprintf("%s: %v", msg, ni.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой.
// Требуется для корректной работы функций errors.Is и errors.As в логгерах и транспортных хендлерах.
func (ni *NotImplementedError) Unwrap() error {
	return ni.err
}
