package errs

import "fmt"

type MiddleWareError struct {
	msg string
	err error
}

var ErrMiddleWare *MiddleWareError

func NewMiddleWareError(msg string, err error) *MiddleWareError {
	return &MiddleWareError{
		msg: msg,
		err: err,
	}
}

func (mwe *MiddleWareError) Error() string {
	msg := "MWARE: fail"
	if mwe.msg != "" {
		msg = fmt.Sprintf("%s: %s", mwe.msg, mwe.err)
	}
	if mwe.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, mwe.err)
	}

	return msg
}

func (mwe *MiddleWareError) Unwrap() error {
	return mwe.err
}
