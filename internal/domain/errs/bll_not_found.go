package errs

import (
	"fmt"
)

type BllNotFoundError struct {
	op    string
	model string
	key   string
	err   error
}

var _ error = (*BllNotFoundError)(nil)

func NewBllNotFoundError(op string, model string, key string, err error) *BllNotFoundError {
	return &BllNotFoundError{
		op:    op,
		model: model,
		key:   key,
		err:   err,
	}
}

func (nf *BllNotFoundError) Error() string {
	msg := fmt.Sprintf("BLL: model %s not found by op %s with key [%s]", nf.model, nf.op, nf.key)
	if nf.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, nf.err)
	}

	return msg
}

func (nf *BllNotFoundError) Unwrap() error {
	return nf.err
}
