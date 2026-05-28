package errs

import (
	"fmt"
)

type ContainerValidateError struct {
	name string
	op   string
	msg  string
	err  error
}

var _ error = (*ContainerError)(nil)

func NewContainerValidateError(name, op, msg string, err error) *ContainerValidateError {
	return &ContainerValidateError{
		name: name,
		op:   op,
		msg:  msg,
		err:  err,
	}
}

func (cve *ContainerValidateError) Error() string {
	msg := "CNT: validate error"
	if cve.name != "" {
		msg = fmt.Sprintf("%s container [%s]", msg, cve.name)
	}
	if cve.op != "" {
		msg = fmt.Sprintf("%s operation [%s]", msg, cve.op)
	}
	if cve.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, cve.msg)
	}
	if cve.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, cve.err)
	}

	return msg
}

func (cve *ContainerValidateError) Unwrap() error {
	return cve.err
}
