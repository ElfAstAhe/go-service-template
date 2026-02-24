package errs

import (
	"fmt"
	"runtime"
)

type DBMigrationError struct {
	msg  string
	err  error
	file string
	line int
}

var ErrDBMigration *DBMigrationError

func NewDBMigrationError(msg string, err error) *DBMigrationError {
	dm := &DBMigrationError{
		msg: msg,
		err: err,
	}

	// runtime.Caller(1) берет данные о том, КТО вызвал NewCommonError
	_, file, line, ok := runtime.Caller(1)
	if ok {
		dm.file = file
		dm.line = line
	}

	return dm
}

func (dm *DBMigrationError) Error() string {
	stack := ""
	if dm.file != "" {
		// Формат [file.go:123] удобен для IDE (можно кликнуть в консоли)
		stack = fmt.Sprintf("[%s:%d] ", dm.file, dm.line)
	}

	msg := "DML: migration error"
	if stack != "" {
		msg = fmt.Sprintf("%s at %s", msg, stack)
	}

	if dm.msg != "" {
		msg = fmt.Sprintf("%s, message: %s", msg, dm.msg)
	}

	if dm.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, dm.err)
	}

	return msg
}

func (dm *DBMigrationError) Unwrap() error {
	return dm.err
}
