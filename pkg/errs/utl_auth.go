package errs

import (
	"fmt"
)

// UtlAuthError реализует нативный интерфейс error, специфицируя низкоуровневые
// инфраструктурные ошибки криптографии, подписи и работы с секретами (Authorization Utility Error).
//
// Инкапсулирует контекст сбоя: пользовательское текстовое описание проблемы (message)
// и ссылку на исходную нижележащую причину падения рантайма (err).
type UtlAuthError struct {
	message string // Человекочитаемое описание сути технического сбоя крипто-хелперов
	err     error  // Ссылка на исходную корневую ошибку (например, ошибку системного генератора случайных чисел)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*UtlAuthError)(nil)

// NewUtlAuthError — фабричный конструктор ошибки утилитарного слоя авторизации.
func NewUtlAuthError(message string, err error) *UtlAuthError {
	return &UtlAuthError{
		message: message,
		err:     err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически выстраивает текстовую строку, четко маркируя сбои подсистем аутентификации и криптографии.
func (ae *UtlAuthError) Error() string {
	msg := "AUTH: util error"
	if ae.message != "" {
		msg = fmt.Sprintf("%s: %s", msg, ae.message)
	}
	if ae.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов крипто-рантайма
		msg = fmt.Sprintf("%s: %v", msg, ae.err)
	}

	return msg
}

// Unwrap возвращает исходную техническую ошибку, инициировавшую сбой.
// Требуется для корректной работы функций errors.Is и errors.As в логгерах и транспортных хендлерах.
func (ae *UtlAuthError) Unwrap() error {
	return ae.err
}
