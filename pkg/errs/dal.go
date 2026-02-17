package errs

import (
	"fmt"
)

// DalError - общая ошибка слоя DAL
type DalError struct {
	op  string
	msg string
	err error
}

var ErrDal *DalError

func NewDalError(op, msg string, err error) *DalError {
	return &DalError{op: op, msg: msg, err: err}
}

func (de *DalError) Error() string {
	msg := "DAL: error"
	if de.op != "" {
		msg = fmt.Sprintf("%s: [%s]", msg, de.op)
	}
	if de.msg != "" {
		msg = fmt.Sprintf("%s %s", msg, de.msg)
	}
	if de.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, de.err)
	}

	return msg
}

func (de *DalError) Unwrap() error {
	return de.err
}
