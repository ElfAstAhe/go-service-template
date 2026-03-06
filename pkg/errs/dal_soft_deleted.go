package errs

import (
	"fmt"
)

type DalSoftDeletedError struct {
	entity string
	key    string
}

var _ error = (*DalSoftDeletedError)(nil)

func NewDalSoftDeletedError(entity string, key string) *DalSoftDeletedError {
	return &DalSoftDeletedError{
		entity: entity,
		key:    key,
	}
}

func (e *DalSoftDeletedError) Error() string {
	return fmt.Sprintf("DAL: %s with key [%s] soft deleted", e.entity, e.key)
}
