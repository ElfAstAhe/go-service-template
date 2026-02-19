package errs

import (
	"fmt"
)

type UtlCipherError struct {
	message string
	err     error
}

var ErrUtlCipher *UtlCipherError

func NewUtlCipherError(msg string, err error) *UtlCipherError {
	return &UtlCipherError{
		message: msg,
		err:     err,
	}
}

func (e *UtlCipherError) Error() string {
	msg := "UTL: cipher error"
	if e.message != "" {
		msg = fmt.Sprintf("%s message [%s]", msg, e.message)
	}
	if e.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

func (e *UtlCipherError) Unwrap() error {
	return e.err
}
