package errs

import (
	"fmt"
)

// TlMappingError реализует нативный интерфейс error, специфицируя ошибки маршалинга,
// десериализации или структурного преобразования данных на транспортном уровне (Transport Layer Mapping Error).
//
// Инкапсулирует расширенный контекст сбоя: имя операции (op), исходный формат/структуру (src),
// целевой формат/структуру (dst), текстовое сообщение (msg) и ссылку на исходную ошибку парсера (err).
type TlMappingError struct {
	op  string // Имя метода или обработчика, где упал маппинг (например, "HTTP.DecodeRequest")
	src string // Наименование источника данных или формата (например, "http_body_json", "grpc_payload")
	dst string // Наименование целевой доменной или транспортной структуры (например, "UserDTO", "CreateLogReq")
	msg string // Человекочитаемое прикладное описание сути технического сбоя
	err error  // Ссылка на исходную корневую ошибку (например, json.SyntaxError или proto.Error)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*TlMappingError)(nil)

// NewTlMappingError — фабричный конструктор ошибки маппинга транспортного уровня.
func NewTlMappingError(
	op string,
	src string,
	dst string,
	msg string,
	err error,
) *TlMappingError {
	return &TlMappingError{
		op:  op,
		src: src,
		dst: dst,
		msg: msg,
		err: err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически выстраивает текстовую цепочку, фиксируя направление маппинга (from src -> to dst) для логов.
func (tme *TlMappingError) Error() string {
	msg := "TL: mapping failed"
	if tme.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, tme.op)
	}
	if tme.src != "" {
		msg = fmt.Sprintf("%s from src %s", msg, tme.src)
	}
	if tme.dst != "" {
		msg = fmt.Sprintf("%s to dst %s", msg, tme.dst)
	}
	if tme.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, tme.msg)
	}
	if tme.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя парсера из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, tme.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку кодировщика или парсера, инициировавшую системный сбой.
// Требуется для корректной работы функций errors.Is и errors.As во внешних логгерах и метриках.
func (tme *TlMappingError) Unwrap() error {
	return tme.err
}
