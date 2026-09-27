package errs

import (
	"fmt"
)

// BllValidateError реализует нативный интерфейс error, специфицируя ошибку нарушения
// бизнес-инвариантов и правил доменной валидации на уровне логики (Domain Layer).
//
// Инкапсулирует контекст падения: имя доменной операции (op), пользовательское текстовое сообщение (msg)
// и ссылку на исходную нижележащую причину сбоя (err).
type BllValidateError struct {
	op  string // Имя метода UseCase/Валидатора, где зафейлилась проверка (например, "ValidateUserAge")
	msg string // Человекочитаемое прикладное описание сути нарушения (например, "user must be at least 18 years old")
	err error  // Ссылка на исходную корневую ошибку (например, ошибку парсинга регулярного выражения)
}

// Гарантируем полное соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*BllValidateError)(nil)

// NewBllValidateError — фабричный конструктор ошибки доменной валидации.
func NewBllValidateError(op string, msg string, err error) *BllValidateError {
	return &BllValidateError{
		op:  op,
		msg: msg,
		err: err,
	}
}

// Error форматирует и возвращает финальное строковое представление ошибки.
// Последовательно склеивает название операции, текст нарушения инварианта и стек вложенных ошибок.
func (bve *BllValidateError) Error() string {
	msg := fmt.Sprintf("BLL: %s validation failed", bve.op)
	if bve.msg != "" {
		msg += ": " + bve.msg
	}

	if bve.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, bve.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой.
// Обеспечивает корректную работу систем автоматического проброса и маппинга ошибок на транспортном слое.
func (bve *BllValidateError) Unwrap() error {
	return bve.err
}
