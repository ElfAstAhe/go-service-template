package http

import (
	"encoding/json"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/transport"
)

func RenderError(rw http.ResponseWriter, err error, errorMapper MapToHTTPStatusFunc) {
	status := errorMapper(err)

	if status >= http.StatusInternalServerError {
		RenderEmpty(rw, status)
	} else {
		RenderJSON(rw, status, transport.NewErrorDTOFromError(status, err), errorMapper)
	}
}

func RenderJSON(rw http.ResponseWriter, status int, data any, errorMapper MapToHTTPStatusFunc) {
	js, err := json.Marshal(data)
	if err != nil {
		RenderError(rw, err, errorMapper)

		return
	}

	rw.Header().Set("Content-Type", MediaTypeApplicationJSON+";charset=utf-8")
	rw.WriteHeader(status)
	_, _ = rw.Write(js)
}

func RenderEmpty(rw http.ResponseWriter, status int) {
	rw.WriteHeader(status)
}
