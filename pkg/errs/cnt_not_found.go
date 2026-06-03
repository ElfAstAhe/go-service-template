package errs

import "fmt"

type ContainerNotFoundError struct {
	msg string
	err error
}

var _ error = (*ContainerNotFoundError)(nil)

func NewContainerNotFoundError(msg string, err error) *ContainerNotFoundError {
	return &ContainerNotFoundError{msg: msg, err: err}
}

func (e *ContainerNotFoundError) Error() string {
	msg := "CNT: not found or not registered"
	if e.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, e.msg)
	}
	if e.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

func (e *ContainerNotFoundError) Unwrap() error {
	return e.err
}
