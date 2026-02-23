package errs

import (
	"fmt"
)

type UtlAuthError struct {
	message string
	err     error
}

var ErrUtlAuth *UtlAuthError

func NewUtlAuthError(message string, err error) *UtlAuthError {
	return &UtlAuthError{
		message: message,
		err:     err,
	}
}

func (ae *UtlAuthError) Error() string {
	msg := "AUTH: util error"
	if ae.message != "" {
		msg = fmt.Sprintf("%s: %s", msg, ae.message)
	}
	if ae.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, ae.err)
	}

	return msg
}

func (ae *UtlAuthError) Unwrap() error {
	return ae.err
}
