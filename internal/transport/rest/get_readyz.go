package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// getReadyz обработчик GET /readyz
func (cr *AppChiRouter) getReadyz(rw http.ResponseWriter, r *http.Request) {
	if !utils.IsNil(cr.readyz) {
		if cr.readyz() {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("READY"))

			return
		}

		rw.WriteHeader(http.StatusServiceUnavailable)
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
