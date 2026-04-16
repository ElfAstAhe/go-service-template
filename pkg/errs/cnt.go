package errs

import (
	"fmt"
)

type ContainerError struct {
	name string
	msg  string
	err  error
}

var _ error = (*ContainerError)(nil)

func NewContainerError(name string, msg string, err error) *ContainerError {
	return &ContainerError{
		name: name,
		msg:  msg,
		err:  err,
	}
}

func (ce *ContainerError) Error() string {
	msg := "CNT: common container error"
	if ce.name != "" {
		msg = fmt.Sprintf("%s, container [%s]", msg, ce.name)
	} else {
		msg = fmt.Sprintf("%s, container [unknown]", msg)
	}
	if ce.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, ce.msg)
	}
	if ce.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, ce.err)
	}

	return msg
}

func (ce *ContainerError) Unwrap() error {
	return ce.err
}
