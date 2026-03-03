package errs

import (
	"fmt"
	"runtime"
)

type CommonError struct {
	msg  string
	err  error
	file string
	line int
}

var _ error = (*CommonError)(nil)

func NewCommonError(msg string, err error) *CommonError {
	e := &CommonError{
		msg: msg,
		err: err,
	}
	// runtime.Caller(1) берет данные о том, КТО вызвал NewCommonError
	_, file, line, ok := runtime.Caller(1)
	if ok {
		e.file = file
		e.line = line
	}

	return e
}

func (ce *CommonError) Error() string {
	stack := ""
	if ce.file != "" {
		// Формат [file.go:123] удобен для IDE (можно кликнуть в консоли)
		stack = fmt.Sprintf("[%s:%d] ", ce.file, ce.line)
	}

	msg := fmt.Sprintf("CMN: error at %s", stack)
	if ce.msg != "" {
		msg = fmt.Sprintf("%s %s", msg, ce.msg)
	}

	if ce.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, ce.err)
	}

	return msg
}

func (ce *CommonError) Unwrap() error {
	return ce.err
}
