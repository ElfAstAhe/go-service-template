package errs

import (
	"fmt"
)

type NotImplementedError struct {
	err error
}

var _ error = (*NotImplementedError)(nil)

func NewNotImplementedError(err error) *NotImplementedError {
	return &NotImplementedError{err: err}
}

func (ni *NotImplementedError) Error() string {
	msg := "CMN: not implemented"
	if ni.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, ni.err)
	}

	return msg
}

func (ni *NotImplementedError) Unwrap() error {
	return ni.err
}
