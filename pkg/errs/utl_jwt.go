package errs

import (
	"fmt"
)

type UtlJWTError struct {
	message string
	err     error
}

var ErrUtlJWT *UtlJWTError

func NewUtlJWTError(msg string, err error) *UtlJWTError {
	return &UtlJWTError{
		message: msg,
		err:     err,
	}
}

func (e *UtlJWTError) Error() string {
	return fmt.Sprintf("jwt util error with message [%s] with error [%v]", e.message, e.err)
}

func (e *UtlJWTError) Unwrap() error {
	return e.err
}
