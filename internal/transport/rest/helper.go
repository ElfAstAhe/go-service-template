package rest

import (
	"encoding/json"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/transport"
	pkghttp "github.com/ElfAstAhe/go-service-template/pkg/transport/http"
)

func (cr *AppChiRouter) renderError(rw http.ResponseWriter, err error) {
	status := mapToHTTPStatus(err)

	if status >= http.StatusInternalServerError {
		cr.renderEmpty(rw, status)
	} else {
		cr.renderJSON(rw, status, transport.NewErrorDTOFromError(status, err))
	}
}

func (cr *AppChiRouter) renderJSON(rw http.ResponseWriter, status int, data any) {
	js, err := json.Marshal(data)
	if err != nil {
		cr.renderError(rw, err)

		return
	}

	rw.Header().Set("Content-Type", pkghttp.MediaTypeApplicationJSON+"; charset=utf-8")
	rw.WriteHeader(status)
	_, _ = rw.Write(js)
}

func (cr *AppChiRouter) renderEmpty(rw http.ResponseWriter, status int) {
	rw.WriteHeader(status)
}
