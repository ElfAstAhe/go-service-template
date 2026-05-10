package http

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
)

func MapToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// 400 BadRequest
	if transport.IsBadRequest(err) {
		return http.StatusBadRequest
	}

	// 401 Unauthorized
	if transport.IsUnauthorized(err) {
		return http.StatusUnauthorized
	}

	// 403 Forbidden
	if transport.IsForbidden(err) {
		return http.StatusForbidden
	}

	// 404 NotFound
	if transport.IsNotFound(err) {
		return http.StatusNotFound
	}

	// 409 Conflict
	if transport.IsConflict(err) {
		return http.StatusConflict
	}

	// 410 Gone
	if transport.IsGone(err) {
		return http.StatusGone
	}

	return http.StatusInternalServerError
}
