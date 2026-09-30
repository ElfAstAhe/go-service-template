package rest

import (
	"net/http"
)

func (cr *AppChiRouter) getHealthz(rw http.ResponseWriter, r *http.Request) {
	if cr.healthz() {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("OK"))

		return
	}

	rw.WriteHeader(http.StatusServiceUnavailable)
}
