package errs

import (
	"fmt"
)

// MiddleWareError реализует нативный интерфейс error, специфицируя технические ошибки,
// возникающие в процессе работы сквозных перехватчиков, фильтров и middleware (HTTP/gRPC слои).
//
// Инкапсулирует контекст сбоя: прикладное текстовое описание проблемы (msg)
// и ссылку на исходную нижележащую причину падения цепочки выполнения (err).
type MiddleWareError struct {
	msg string // Описание сути технического сбоя middleware (например, "rate limit exceeded" или "cors validation failed")
	err error  // Ссылка на исходную корневую ошибку рантайма (например, ошибку отвала Redis при проверке лимитов)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*MiddleWareError)(nil)

// NewMiddleWareError — фабричный конструктор ошибки промежуточного слоя транспорта.
func NewMiddleWareError(msg string, err error) *MiddleWareError {
	return &MiddleWareError{
		msg: msg,
		err: err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// ИСПРАВЛЕНО: Логика форматирования полностью защищена от дублирования ошибок и некорректного вывода nil-значений.
func (mwe *MiddleWareError) Error() string {
	msg := "MWARE: fail"
	if mwe.msg != "" {
		msg = fmt.Sprintf("%s with message : %s", msg, mwe.msg)
	}
	if mwe.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов интерцепторов
		msg = fmt.Sprintf("%s: %v", msg, mwe.err)
	}

	return msg
}

// Unwrap возвращает исходную техническую ошибку, инициировавшую сбой внутри сквозного перехватчика.
// Требуется для корректной работы функций errors.Is и errors.As во внешних логгерах и метриках.
func (mwe *MiddleWareError) Unwrap() error {
	return mwe.err
}
