package errs

import (
	"fmt"
)

type ConfigError struct {
	msg string
	err error
}

var ErrConfig *ConfigError

// Гарантируем соответствие интерфейсу на этапе компиляции
var _ error = (*ConfigError)(nil)

func NewConfigError(msg string, err error) *ConfigError {
	return &ConfigError{
		msg: msg,
		err: err,
	}
}

func (e *ConfigError) Error() string {
	msg := "CFG: error"
	if e.msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.msg)
	}
	if e.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

func (e *ConfigError) Unwrap() error {
	return e.err
}
