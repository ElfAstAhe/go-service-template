package rest

import (
	"net/http"
)

func (cr *AppChiRouter) getReadyz(rw http.ResponseWriter, r *http.Request) {
	if cr.readyz() {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("READY"))

		return
	}

	rw.WriteHeader(http.StatusServiceUnavailable)
}
