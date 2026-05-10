package errs

import (
	"fmt"
)

type BllError struct {
	op  string
	msg string
	err error
}

var _ error = (*BllError)(nil)

func NewBllError(op string, msg string, err error) *BllError {
	return &BllError{
		op:  op,
		msg: msg,
		err: err,
	}
}

func (e *BllError) Error() string {
	msg := "BLL: error"
	if e.op != "" {
		msg += " " + e.op
	}
	if e.msg != "" {
		msg += ": " + e.msg
	}
	if e.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

func (e *BllError) Unwrap() error {
	return e.err
}
