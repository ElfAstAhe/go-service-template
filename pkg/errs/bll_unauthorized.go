package errs

import (
	"fmt"
)

type BllUnauthorizedError struct {
	op  string
	msg string
	err error
}

var _ error = (*BllUnauthorizedError)(nil)

func NewBllUnauthorizedError(op string, msg string, err error) *BllUnauthorizedError {
	return &BllUnauthorizedError{op: op, msg: msg, err: err}
}

func (beu *BllUnauthorizedError) Error() string {
	msg := "BLL: unauthorized"
	if beu.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, beu.op)
	}
	if beu.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, beu.msg)
	}
	if beu.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, beu.err)
	}

	return msg
}

func (beu *BllUnauthorizedError) Unwrap() error {
	return beu.err
}
