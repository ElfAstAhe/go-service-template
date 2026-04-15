package errs

import "fmt"

type DalCacheError struct {
	op  string
	msg string
	err error
}

var _ error = (*DalCacheError)(nil)

func NewDalCacheError(op string, msg string, err error) *DalCacheError {
	return &DalCacheError{
		op:  op,
		err: err,
	}
}

func (dc *DalCacheError) Error() string {
	msg := "DAL: accessing cache failed"
	if dc.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, dc.op)
	}
	if dc.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, dc.msg)
	}
	if dc.err != nil {
		msg = fmt.Sprintf("%s: with error %v", msg, dc.err)
	}

	return msg
}

func (dc *DalCacheError) Unwrap() error {
	return dc.err
}
