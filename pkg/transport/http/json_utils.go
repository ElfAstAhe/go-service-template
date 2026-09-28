package http

import (
	"encoding/json"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// DecodeJSON осуществляет безопасное, строго типизированное декодирование входящего тела
// HTTP-запроса в целевую доменную структуру или объект передачи данных (DTO).
//
// 🛡️ Защитные механизмы рантайма:
// 1. Ограничивает чтение буфера сокета лимитом в 1 МБ для пресечения DoS-атак переполнения RAM.
// 2. Блокирует обработку при наличии неизвестных полей (Strict mode) для жесткого контроля контрактов API.
// 3. Автоматически упаковывает сбои парсинга в TlMappingError с фиксацией абсолютных путей типов.
func DecodeJSON(r *http.Request, dst any) error {
	// 1. Ограничиваем чтение (например, 1Мб), чтобы не выесть RAM.
	// MaxBytesReader автоматически вернет ошибку, если тело больше установленного лимита.
	r.Body = http.MaxBytesReader(nil, r.Body, 1024*1024)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	// 2. Strict mode: если клиент прислал поле, которого нет в DTO — это 400.
	// Помогает гарантированно отловить опечатки на фронте (например, "iddd" вместо "id").
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return errs.NewTlMappingError("DecodeJSON", utils.GetFullTypeName(r), utils.GetFullTypeName(dst), "decode JSON", err)
	}

	return nil
}
