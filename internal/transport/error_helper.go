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

func IsNotFound(err error) bool {
	var (
		errBllNotFound *errs.BllNotFoundError
	)

	return errors.As(err, &errBllNotFound)
}

func IsConflict(err error) bool {
	var (
		errBllUnique *errs.BllUniqueError
	)

	return errors.As(err, &errBllUnique)
}
