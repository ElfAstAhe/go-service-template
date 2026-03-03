package errs

import (
	"fmt"
)

type UtlJWTError struct {
	message string
	err     error
}

var _ error = (*UtlJWTError)(nil)

func NewUtlJWTError(msg string, err error) *UtlJWTError {
	return &UtlJWTError{
		message: msg,
		err:     err,
	}
}

func (e *UtlJWTError) Error() string {
	msg := "UTL: jwt error"
	if e.message != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.message)
	}
	if e.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

func (e *UtlJWTError) Unwrap() error {
	return e.err
}
