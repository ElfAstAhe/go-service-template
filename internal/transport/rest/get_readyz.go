package rest

import (
	"net/http"
)

// getReadyz обработчик GET /readyz
func (cr *AppChiRouter) getReadyz(rw http.ResponseWriter, r *http.Request) {
	if cr.readyz() {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("READY"))

		return
	}

	rw.WriteHeader(http.StatusServiceUnavailable)
}
