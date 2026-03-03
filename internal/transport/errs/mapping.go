package errs

import (
	"fmt"
)

type TrMappingError struct {
	op  string
	src string
	dst string
	msg string
	err error
}

var _ error = (*TrMappingError)(nil)

func NewTrMappingError(op string, src string, dst string, msg string, err error) *TrMappingError {
	return &TrMappingError{op, src, dst, msg, err}
}

func (tme *TrMappingError) Error() string {
	msg := fmt.Sprintf("TR: %s mapping failed", tme.op)
	if tme.src != "" {
		msg = fmt.Sprintf("%s from src %s", msg, tme.src)
	}
	if tme.dst != "" {
		msg = fmt.Sprintf("%s to dst %s", msg, tme.dst)
	}
	if tme.msg != "" {
		msg = fmt.Sprintf("%s msg %s", msg, tme.msg)
	}
	if tme.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, tme.err)
	}

	return msg
}

func (tme *TrMappingError) Unwrap() error {
	return tme.err
}
