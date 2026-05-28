package errs

import (
	"fmt"
)

// TlCommonError transport layer common error
type TlCommonError struct {
	op  string
	msg string
	err error
}

func NewTlCommonError(op string, msg string, err error) *TlCommonError {
	return &TlCommonError{
		op:  op,
		msg: msg,
		err: err,
	}
}

func (tce *TlCommonError) Error() string {
	msg := "TL: common error"
	if tce.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, tce.op)
	}
	if tce.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, tce.msg)
	}
	if tce.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, tce.err)
	}

	return msg
}

func (tce *TlCommonError) Unwrap() error {
	return tce.err
}
