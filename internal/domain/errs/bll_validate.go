package errs

import "fmt"

type BllValidateError struct {
	op  string
	msg string
	err error
}

var ErrBllValidate *BllValidateError

func NewBllValidateError(op string, msg string, err error) *BllValidateError {
	return &BllValidateError{
		op:  op,
		msg: msg,
		err: err,
	}
}

func (bve *BllValidateError) Error() string {
	msg := fmt.Sprintf("BLL: %s validation failed", bve.op)
	if bve.msg != "" {
		msg += ": " + bve.msg
	}

	if bve.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, bve.err)
	}

	return msg
}

func (bve *BllValidateError) Unwrap() error {
	return bve.err
}
