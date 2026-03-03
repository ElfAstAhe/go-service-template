package transport

import (
	"errors"

	domerrs "github.com/ElfAstAhe/go-service-template/internal/domain/errs"
	transperrs "github.com/ElfAstAhe/go-service-template/internal/transport/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

func IsBadRequest(err error) bool {
	var (
		errInvalidArgument *errs.InvalidArgumentError
		errBllValidate     *domerrs.BllValidateError
		errTrMapping       *transperrs.TrMappingError
	)

	return errors.As(err, &errInvalidArgument) ||
		errors.As(err, &errBllValidate) ||
		errors.As(err, &errTrMapping)
}

func IsNotFound(err error) bool {
	var (
		errBllNotFound *domerrs.BllNotFoundError
	)

	return errors.As(err, &errBllNotFound)
}

func IsConflict(err error) bool {
	var (
		errBllUnique *domerrs.BllUniqueError
	)

	return errors.As(err, &errBllUnique)
}
