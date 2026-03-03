package errs

import (
	"fmt"
)

// InvalidArgumentError - некорректный аргумент, приводт к Bad Request
type InvalidArgumentError struct {
	param string
	value any
	err   error
}

var _ error = (*InvalidArgumentError)(nil)

func NewInvalidArgumentError(param string, value any) *InvalidArgumentError {
	return NewInvalidArgumentErrorChain(param, value, nil)
}

func NewInvalidArgumentErrorChain(param string, value any, err error) *InvalidArgumentError {
	return &InvalidArgumentError{param: param, value: value, err: err}
}

func (e *InvalidArgumentError) Error() string {
	msg := fmt.Sprintf("CMN: invalid argument [%s] with value [%v]", e.param, e.value)
	if e.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

func (e *InvalidArgumentError) Unwrap() error {
	return e.err
}
