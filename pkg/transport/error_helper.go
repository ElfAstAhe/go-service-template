package transport

import (
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

func IsBadRequest(err error) bool {
	var (
		errInvalidArgument *errs.InvalidArgumentError
		errBllValidate     *errs.BllValidateError
		errTrMapping       *errs.TlMappingError
	)

	return errors.As(err, &errInvalidArgument) ||
		errors.As(err, &errBllValidate) ||
		errors.As(err, &errTrMapping)
}

func IsUnauthorized(err error) bool {
	var (
		errBllUnauthorized *errs.BllUnauthorizedError
	)

	return errors.As(err, &errBllUnauthorized)
}

func IsForbidden(err error) bool {
	var (
		errBllForbidden *errs.BllForbiddenError
	)

	return errors.As(err, &errBllForbidden)
}

func IsNotFound(err error) bool {
	var (
		errBllNotFound *errs.BllNotFoundError
		errDalNotFound *errs.DalNotFoundError
	)

	return errors.As(err, &errBllNotFound) ||
		errors.As(err, &errDalNotFound)
}

func IsConflict(err error) bool {
	var (
		errBllUnique        *errs.BllUniqueError
		errDalAlreadyExists *errs.DalAlreadyExistsError
	)

	return errors.As(err, &errBllUnique) ||
		errors.As(err, &errDalAlreadyExists)
}

func IsGone(err error) bool {
	var (
		errDalSoftDeleted *errs.DalSoftDeletedError
	)

	return errors.As(err, &errDalSoftDeleted)
}
