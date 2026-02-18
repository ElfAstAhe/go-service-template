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

var ErrCommon *CommonError

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

	if ce.err != nil {
		return fmt.Sprintf("%sBLL: %s: %v", stack, ce.msg, ce.err)
	}

	return fmt.Sprintf("%sCMN: %s", stack, ce.msg)
}

func (ce *CommonError) Unwrap() error {
	return ce.err
}
