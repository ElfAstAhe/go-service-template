package errs

import (
	"fmt"
)

type TlMappingError struct {
	op  string
	src string
	dst string
	msg string
	err error
}

var _ error = (*TlMappingError)(nil)

func NewTlMappingError(
	op string,
	src string,
	dst string,
	msg string,
	err error,
) *TlMappingError {
	return &TlMappingError{
		op:  op,
		src: src,
		dst: dst,
		msg: msg,
		err: err,
	}
}

func (tme *TlMappingError) Error() string {
	msg := "TL: mapping failed"
	if tme.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, tme.op)
	}
	if tme.src != "" {
		msg = fmt.Sprintf("%s from src %s", msg, tme.src)
	}
	if tme.dst != "" {
		msg = fmt.Sprintf("%s to dst %s", msg, tme.dst)
	}
	if tme.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, tme.msg)
	}
	if tme.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, tme.err)
	}

	return msg
}

func (tme *TlMappingError) Unwrap() error {
	return tme.err
}
