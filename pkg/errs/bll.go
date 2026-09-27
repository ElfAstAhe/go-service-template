package errs

import (
	"fmt"
)

// BllError реализует нативный интерфейс error, представляя собой базовую системную ошибку
// слоя бизнес-логики (Business Logic Layer Error).
//
// Инкапсулирует контекст падения: имя доменной операции (op), пользовательское текстовое сообщение (msg)
// и ссылку на исходную нижележащую причину сбоя (err) для обеспечения сквозного оборачивания ошибок.
type BllError struct {
	op  string // Имя выполняемой бизнес-операции или метода UseCase (например, "Login", "RegisterUser")
	msg string // Человекочитаемое прикладное описание сути проблемы
	err error  // Ссылка на исходную корневую ошибку (может быть ошибкой DAL-слоя или сети)
}

// Гарантируем полное соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*BllError)(nil)

// NewBllError — фабричный конструктор ошибки бизнес-логики.
func NewBllError(op string, msg string, err error) *BllError {
	return &BllError{
		op:  op,
		msg: msg,
		err: err,
	}
}

// Error форматирует и возвращает финальное строковое представление ошибки.
// Динамически собирает текстовый паспорт сбоя с разграничением сегментов двоеточиями.
func (e *BllError) Error() string {
	msg := "BLL: error"
	if e.op != "" {
		msg += " " + e.op
	}
	if e.msg != "" {
		msg += ": " + e.msg
	}
	if e.err != nil {
		// Обогащаем строку выводом нижележащей ошибки из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой.
// Является фундаментальным контрактом рантайма Go для поддержки работы функций errors.Is и errors.As.
func (e *BllError) Unwrap() error {
	return e.err
}
