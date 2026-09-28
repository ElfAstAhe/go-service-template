package errs

import (
	"fmt"
)

// InvalidArgumentError реализует нативный интерфейс error, специфицируя системную ошибку
// передачи некорректного, невалидного или пустого входящего параметра (Bad Request).
//
// Инкапсулирует структурированный контекст сбоя: имя невалидного параметра (param),
// само переданное ошибочное значение (value) и ссылку на исходную ошибку валидации (err).
type InvalidArgumentError struct {
	param string // Имя аргумента, не прошедшего валидацию (например, "limit", "id")
	value any    // Фактическое невалидное значение любого типа (например, -1, "invalid-uuid")
	err   error  // Ссылка на исходную системную ошибку парсинга или проверки (опционально)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*InvalidArgumentError)(nil)

// NewInvalidArgumentError — фабричный конструктор ошибки невалидного аргумента (без вложенной ошибки).
func NewInvalidArgumentError(param string, value any) *InvalidArgumentError {
	return NewInvalidArgumentErrorChain(param, value, nil)
}

// NewInvalidArgumentErrorChain — расширенный фабричный конструктор с поддержкой сквозного связывания ошибок (Error Chain).
func NewInvalidArgumentErrorChain(param string, value any, err error) *InvalidArgumentError {
	return &InvalidArgumentError{param: param, value: value, err: err}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Формирует строго структурированное сообщение, фиксируя имя аргумента и его ошибочное значение.
func (e *InvalidArgumentError) Error() string {
	msg := fmt.Sprintf("CMN: invalid argument [%s] with value [%v]", e.param, e.value)
	if e.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку валидации или парсинга, инициировавшую сбой.
// Требуется для корректной работы функций errors.Is и errors.As во внешних классификаторах транспорта.
func (e *InvalidArgumentError) Unwrap() error {
	return e.err
}
