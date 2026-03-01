package errs

import (
	"fmt"
)

type BllUniqueError struct {
	op    string
	model string
	key   string
	err   error
}

var _ error = (*BllUniqueError)(nil)

func NewBllUniqueError(op string, model string, key string, err error) *BllUniqueError {
	return &BllUniqueError{
		op:    op,
		model: model,
		key:   key,
		err:   err,
	}
}

func (u *BllUniqueError) Error() string {
	msg := fmt.Sprintf("BLL: %s model already exists, op %s", u.model, u.op)
	if u.key != "" {
		msg = fmt.Sprintf("%s key [%s]", msg, u.key)
	}
	if u.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, u.err)
	}

	return msg
}

func (u *BllUniqueError) Unwrap() error {
	return u.err
}
