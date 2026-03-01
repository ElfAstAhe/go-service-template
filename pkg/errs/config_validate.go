package errs

import (
	"fmt"
)

type ConfigValidateError struct {
	group string
	item  string
	msg   string
	err   error
}

// Гарантируем соответствие интерфейсу на этапе компиляции
var _ error = (*ConfigValidateError)(nil)

func NewConfigValidateError(group string, item string, msg string, err error) *ConfigValidateError {
	return &ConfigValidateError{
		group: group,
		item:  item,
		msg:   msg,
		err:   err,
	}
}

func (cv *ConfigValidateError) Error() string {
	msg := "CFG: validate error"
	if cv.group != "" {
		msg = fmt.Sprintf("%s: group %s", msg, cv.group)
	}
	if cv.item != "" {
		msg = fmt.Sprintf("%s: item %s", msg, cv.item)
	}
	if cv.msg != "" {
		msg = fmt.Sprintf("%s: msg %s", msg, cv.msg)
	}
	if cv.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, cv.err)
	}

	return msg
}

func (cv *ConfigValidateError) Unwrap() error {
	return cv.err
}
