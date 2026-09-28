package errs

import (
	"fmt"
)

// UtlJWTError реализует нативный интерфейс error, специфицируя низкоуровневые
// технические ошибки парсинга, валидации подписей и обработки токенов JWT (JWT Utility Error).
//
// Инкапсулирует контекст сбоя: пользовательское текстовое описание проблемы (message)
// и ссылку на исходную нижележащую причину падения токен-рантайма (err).
type UtlJWTError struct {
	message string // Человекочитаемое описание сути технического сбоя токенизатора
	err     error  // Ссылка на исходную корневую ошибку (например, ошибку валидации подписи от библиотеки golang-jwt)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*UtlJWTError)(nil)

// NewUtlJWTError — фабричный конструктор ошибки JWT-слоя.
func NewUtlJWTError(msg string, err error) *UtlJWTError {
	return &UtlJWTError{
		message: msg,
		err:     err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически выстраивает текстовую строку, четко маркируя сбои подсистем токенизации и авторизации.
func (e *UtlJWTError) Error() string {
	msg := "UTL: jwt error"
	if e.message != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.message)
	}
	if e.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов библиотеки
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

// Unwrap возвращает исходную техническую ошибку, инициировавшую сбой.
// Требуется для корректной работы функций errors.Is и errors.As в логгерах и транспортных хендлерах.
func (e *UtlJWTError) Unwrap() error {
	return e.err
}
