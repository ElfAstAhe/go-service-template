package errs

import (
	"fmt"
)

// DalAlreadyExistsError — ошибка уникальности данные
type DalAlreadyExistsError struct {
	Entity string // Какая сущность (например, "User" или "UserData")
	Value  any    // Какое значение вызвало конфликт (например, "login 'admin'")
	Err    error  // Исходная ошибка из драйвера БД (опционально)
}

var ErrDalAlreadyExists *DalAlreadyExistsError

func NewDalAlreadyExistsError(entity string, value any, err error) *DalAlreadyExistsError {
	return &DalAlreadyExistsError{
		Entity: entity,
		Value:  value,
		Err:    err,
	}
}

func (e *DalAlreadyExistsError) Error() string {
	msg := fmt.Sprintf("DAL: %s with value [%v] already exists", e.Entity, e.Value)
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", msg, e.Err)
	}

	return msg
}

func (e *DalAlreadyExistsError) Unwrap() error {
	return e.Err
}
