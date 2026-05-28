package errs

import (
	"fmt"
)

type BllForbiddenError struct {
	op  string
	msg string
	err error
}

var _ error = (*BllForbiddenError)(nil)

func NewBllForbiddenError(op string, msg string, err error) *BllForbiddenError {
	return &BllForbiddenError{op, msg, err}
}

func (bfe *BllForbiddenError) Error() string {
	msg := "BLL: forbidden"
	if bfe.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, bfe.op)
	}
	if bfe.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, bfe.msg)
	}
	if bfe.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, bfe.err)
	}

	return msg
}

func (bfe *BllForbiddenError) Unwrap() error {
	return bfe.err
}
