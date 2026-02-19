package errs

import (
	"fmt"
)

// InvalidArgumentError - некорректный аргумент, приводт к Bad Request
type InvalidArgumentError struct {
	Param string
	Value any
}

var ErrAppInvalidArgument *InvalidArgumentError

func NewInvalidArgumentError(param string, value any) *InvalidArgumentError {
	return &InvalidArgumentError{Param: param, Value: value}
}

func (e *InvalidArgumentError) Error() string {
	return fmt.Sprintf("CMN: invalid argument [%s] with value [%v]", e.Param, e.Value)
}
