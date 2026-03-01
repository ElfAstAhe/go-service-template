package rest

import (
	"errors"
	"net/http"

	domerrs "github.com/ElfAstAhe/go-service-template/internal/domain/errs"
	transperrs "github.com/ElfAstAhe/go-service-template/internal/transport/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

func mapToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	// 400 BadRequest
	if isBadRequest(err) {
		return http.StatusBadRequest
	}

	// 404 NotFound
	if isNotFound(err) {
		return http.StatusNotFound
	}

	// 409 Conflict
	if isConflict(err) {
		return http.StatusConflict
	}

	return http.StatusInternalServerError
}

func isBadRequest(err error) bool {
	var (
		errInvalidArgument *errs.InvalidArgumentError
		errBllValidate     *domerrs.BllValidateError
		errTrMapping       *transperrs.TrMappingError
	)

	return errors.As(err, &errInvalidArgument) ||
		errors.As(err, &errBllValidate) ||
		errors.As(err, &errTrMapping)
}

func isNotFound(err error) bool {
	var (
		errBllNotFound *domerrs.BllNotFoundError
	)

	return errors.As(err, &errBllNotFound)
}

func isConflict(err error) bool {
	var (
		errBllUnique *domerrs.BllUniqueError
	)

	return errors.As(err, &errBllUnique)
}
