package errs

import (
	"fmt"
)

// BllUnauthorizedError реализует нативный интерфейс error, специфицируя ошибку нарушения
// или отсутствия аутентификации (невалидный сессионный токен, неверный пароль) на уровне домена (Domain Layer).
//
// Инкапсулирует контекст падения: имя доменной операции (op), пользовательское текстовое сообщение (msg)
// и ссылку на исходную нижележащую причину сбоя (err).
type BllUnauthorizedError struct {
	op  string // Имя метода UseCase, где пользователь не прошёл проверку подлинности (например, "ChangePassword")
	msg string // Человекочитаемое прикладное описание проблемы (например, "session token has expired")
	err error  // Ссылка на исходную корневую ошибку (например, ошибку валидации из библиотеки golang-jwt)
}

// Гарантируем полное соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*BllUnauthorizedError)(nil)

// NewBllUnauthorizedError — фабричный конструктор ошибки аутентификации.
func NewBllUnauthorizedError(op string, msg string, err error) *BllUnauthorizedError {
	return &BllUnauthorizedError{op: op, msg: msg, err: err}
}

// Error форматирует и возвращает финальное строковое представление ошибки.
// Последовательно склеивает контекст доменной операции, текстовое сообщение и стек вложенных ошибок.
func (beu *BllUnauthorizedError) Error() string {
	msg := "BLL: unauthorized"
	if beu.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, beu.op)
	}
	if beu.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, beu.msg)
	}
	if beu.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, beu.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой.
// Обеспечивает корректную работу систем автоматического проброса и маппинга ошибок на транспортном слое.
func (beu *BllUnauthorizedError) Unwrap() error {
	return beu.err
}
