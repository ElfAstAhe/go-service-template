package rest

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/transport"
)

func mapToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// 400 BadRequest
	if transport.IsBadRequest(err) {
		return http.StatusBadRequest
	}

	// 404 NotFound
	if transport.IsNotFound(err) {
		return http.StatusNotFound
	}

	// 409 Conflict
	if transport.IsConflict(err) {
		return http.StatusConflict
	}

	return http.StatusInternalServerError
}
