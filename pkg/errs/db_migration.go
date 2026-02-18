package errs

import (
	"fmt"
)

type DBMigrationError struct {
	migration string
	err       error
}

var ErrDBMigration *DBMigrationError

func NewDBMigrationError(migration string, err error) *DBMigrationError {
	return &DBMigrationError{
		migration: migration,
		err:       err,
	}
}

func (dm *DBMigrationError) Error() string {
	msg := "DB Migration: failed"
	if dm.migration != "" {
		msg = fmt.Sprintf("%s: migration [%s]", msg, dm.migration)
	}
	if dm.err != nil {
		msg = fmt.Sprintf("%s: %v", msg, dm.err)
	}

	return msg
}

func (dm *DBMigrationError) Unwrap() error {
	return dm.err
}
