package errs

import (
	"fmt"
)

type UtilsError struct {
	op  string
	msg string
	err error
}

var ErrUtils *UtilsError

func NewUtilsError(op, msg string, err error) *UtilsError {
	return &UtilsError{
		op:  op,
		msg: msg,
		err: err,
	}
}

func (ue *UtilsError) Error() string {
	msg := fmt.Sprintf("UTL: %s error", ue.op)
	if ue.msg != "" {
		msg = fmt.Sprintf("%s %s", msg, ue.msg)
	}
	if ue.err != nil {
		msg = fmt.Sprintf("%s %v", msg, ue.err)
	}

	return msg
}

func (ue *UtilsError) Unwrap() error {
	return ue.err
}
