package errs

import (
	"fmt"
)

// DalSoftDeletedError реализует нативный интерфейс error, специфицируя ситуацию,
// когда запрашиваемый ресурс был ранее логически удален (Soft Delete) на уровне физической СУБД.
//
// Инкапсулирует контекст сбоя: название целевой доменной сущности (entity) и её поисковый ключ (key).
type DalSoftDeletedError struct {
	entity string // Имя мягко удаленной доменной сущности или таблицы (например, "User", "Article")
	key    string // Строковое представление первичного ключа или UUID, по которому производился поиск
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*DalSoftDeletedError)(nil)

// NewDalSoftDeletedError — фабричный конструктор ошибки обращения к логически удаленной записи.
func NewDalSoftDeletedError(entity string, key string) *DalSoftDeletedError {
	return &DalSoftDeletedError{
		entity: entity,
		key:    key,
	}
}

// Error форматирует и возвращает лаконичный строковый паспорт ошибки для систем логирования.
func (e *DalSoftDeletedError) Error() string {
	return fmt.Sprintf("DAL: %s with key [%s] soft deleted", e.entity, e.key)
}
