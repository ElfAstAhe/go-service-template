package http

import (
	"encoding/json"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

func DecodeJSON(r *http.Request, dst any) error {
	// 1. Ограничиваем чтение (например, 1Мб), чтобы не выесть RAM
	// MaxBytesReader вернет ошибку, если тело больше лимита
	r.Body = http.MaxBytesReader(nil, r.Body, 1024*1024)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	// 2. Strict mode: если клиент прислал поле, которого нет в DTO — это 400
	// Помогает отловить опечатки на фронте (например, "iddd" вместо "id")
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return errs.NewTlMappingError("DecodeJSON", utils.GetFullTypeName(r), utils.GetFullTypeName(dst), "decode JSON", err)
	}

	return nil
}
