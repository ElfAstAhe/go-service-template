package errs

import (
	"fmt"
)

type UtlError struct {
	op  string
	msg string
	err error
}

var _ error = (*UtlError)(nil)

func NewUtlError(op, msg string, err error) *UtlError {
	return &UtlError{
		op:  op,
		msg: msg,
		err: err,
	}
}

func (ue *UtlError) Error() string {
	msg := fmt.Sprintf("UTL: %s error", ue.op)
	if ue.msg != "" {
		msg = fmt.Sprintf("%s %s", msg, ue.msg)
	}
	if ue.err != nil {
		msg = fmt.Sprintf("%s %v", msg, ue.err)
	}

	return msg
}

func (ue *UtlError) Unwrap() error {
	return ue.err
}
