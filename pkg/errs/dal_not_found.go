package errs

import (
	"fmt"
)

// DalNotFoundError — отсутствие сущности
type DalNotFoundError struct {
	Entity string // Какая сущность (например, "User" или "UserData")
	Value  any    // Какое значение вызвало конфликт (например, "login 'admin'")
	Err    error  // Исходная ошибка из драйвера БД (опционально)
}

var ErrDalNotFound *DalNotFoundError

func NewDalNotFoundError(entity string, value any, err error) *DalNotFoundError {
	return &DalNotFoundError{
		Entity: entity,
		Value:  value,
		Err:    err,
	}
}

func (dnf *DalNotFoundError) Error() string {
	msg := fmt.Sprintf("DAL: %s with value [%v] not found", dnf.Entity, dnf.Value)
	if dnf.Err != nil {
		return fmt.Sprintf("%s: %v", msg, dnf.Err)
	}

	return msg
}

func (dnf *DalNotFoundError) Unwrap() error {
	return dnf.Err
}
