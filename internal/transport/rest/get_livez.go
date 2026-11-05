package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// getHealthz обработчик пути GET /healthz
func (cr *AppChiRouter) getLivez(rw http.ResponseWriter, r *http.Request) {
	if !utils.IsNil(cr.livez) {
		if cr.livez() {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("OK"))

			return
		}

		rw.WriteHeader(http.StatusServiceUnavailable)
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
