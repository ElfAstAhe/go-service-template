package errs

import (
	"fmt"
)

// DalAlreadyExistsError реализует нативный интерфейс error, специфицируя ошибку нарушения
// уникального индекса (Unique Violation / Коллизия данных) на уровне физической базы данных.
//
// Инкапсулирует расширенный контекст: название доменной сущности (Entity), конкретное значение
// составного или первичного ключа, давшее конфликт (Value), и ссылку на исходную ошибку драйвера СУБД.
type DalAlreadyExistsError struct {
	Entity string // Имя доменной сущности/таблицы, где обнаружен дубликат (например, "User", "Client")
	Value  any    // Значение или структура, вызвавшая коллизию (например, "email 'test@test.ru'")
	Err    error  // Исходная системная ошибка из драйвера БД (например, pq.Error со значением кода 23505)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*DalAlreadyExistsError)(nil)

// NewDalAlreadyExistsError — фабричный конструктор ошибки дублирования данных на уровне DAL.
func NewDalAlreadyExistsError(entity string, value any, err error) *DalAlreadyExistsError {
	return &DalAlreadyExistsError{
		Entity: entity,
		Value:  value,
		Err:    err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Формирует строго структурированное сообщение, фиксируя целевую сущность и конфликтующее значение ключа.
func (e *DalAlreadyExistsError) Error() string {
	msg := fmt.Sprintf("DAL: %s with value [%v] already exists", e.Entity, e.Value)
	if e.Err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов драйвера
		return fmt.Sprintf("%s: %v", msg, e.Err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку драйвера СУБД, инициировавшую сбой при проверке ограничений (Constraints).
// Требуется для корректной работы функций errors.Is и errors.As во внешних классификаторах транспорта.
func (e *DalAlreadyExistsError) Unwrap() error {
	return e.Err
}
