package errs

import (
	"fmt"
)

// UtlCipherError реализует нативный интерфейс error, специфицируя низкоуровневые
// технические ошибки шифрования, дешифрования и валидации целостности данных (Cipher Utility Error).
//
// Инкапсулирует контекст сбоя: пользовательское текстовое описание проблемы (message)
// и ссылку на исходную нижележащую причину падения крипто-рантайма (err).
type UtlCipherError struct {
	message string // Человекочитаемое описание сути технического сбоя шифратора
	err     error  // Ссылка на исходную корневую ошибку (например, ошибку валидации тега аутентификации GCM)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*UtlCipherError)(nil)

// NewUtlCipherError — фабричный конструктор ошибки криптографического слоя.
func NewUtlCipherError(msg string, err error) *UtlCipherError {
	return &UtlCipherError{
		message: msg,
		err:     err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически выстраивает текстовую строку, четко маркируя сбои подсистем симметричного/асимметричного шифрования.
func (e *UtlCipherError) Error() string {
	msg := "UTL: cipher error"
	if e.message != "" {
		msg = fmt.Sprintf("%s message [%s]", msg, e.message)
	}
	if e.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов крипто-рантайма
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

// Unwrap возвращает исходную техническую ошибку, инициировавшую сбой.
// Требуется для корректной работы функций errors.Is и errors.As в логгерах и транспортных хендлерах.
func (e *UtlCipherError) Unwrap() error {
	return e.err
}
