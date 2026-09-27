package errs

import (
	"fmt"
)

// TlCommonError реализует нативный интерфейс error, специфицируя базовые системные
// и инфраструктурные сбои общего характера на транспортном уровне (Transport Layer Common Error).
//
// Инкапсулирует контекст падения: имя сетевой/транспортной операции (op), пользовательское текстовое
// описание проблемы (msg) и ссылку на исходную корневую причину сетевого сбоя (err).
type TlCommonError struct {
	op  string // Имя сетевого метода, хендлера или интерцептора (например, "gRPC.UnaryServerInterceptor")
	msg string // Человекочитаемое описание сути технического сбоя сетевого или брокерского протокола
	err error  // Ссылка на исходную корневую ошибку (например, системную ошибку сетевого сокета ОС)
}

var _ error = (*TlCommonError)(nil)

// NewTlCommonError — фабричный конструктор общей транспортной ошибки.
func NewTlCommonError(op string, msg string, err error) *TlCommonError {
	return &TlCommonError{
		op:  op,
		msg: msg,
		err: err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Последовательно склеивает контекст транспортной операции, текстовое сообщение и стек вложенных ошибок.
func (tce *TlCommonError) Error() string {
	msg := "TL: common error"
	if tce.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, tce.op)
	}
	if tce.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, tce.msg)
	}
	if tce.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сетевого сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, tce.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку сетевого драйвера или протокола, инициировавшую сбой.
// Требуется для корректной работы функций errors.Is и errors.As во внешних логгерах и метриках.
func (tce *TlCommonError) Unwrap() error {
	return tce.err
}
