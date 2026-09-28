package errs

import (
	"fmt"
)

// DalNotFoundError реализует нативный интерфейс error, специфицируя ошибку физического
// отсутствия запрашиваемой записи или строки (sql.ErrNoRows) на уровне базы данных (DAL слой).
//
// Инкапсулирует структурированный контекст сбоя: название целевой сущности/таблицы (Entity),
// значение или критерий поиска, по которому ничего не нашли (Value), и ссылку на исходную ошибку драйвера СУБД.
type DalNotFoundError struct {
	Entity string // Имя доменной сущности или физической таблицы (например, "User", "AuditLog")
	Value  any    // Идентификатор или критерий, по которому выполнялся поиск (например, конкретный UUID)
	Err    error  // Исходная системная ошибка из драйвера БД (опционально, например sql.ErrNoRows)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*DalNotFoundError)(nil)

// NewDalNotFoundError — фабричный конструктор ошибки отсутствия записи на уровне DAL.
func NewDalNotFoundError(entity string, value any, err error) *DalNotFoundError {
	return &DalNotFoundError{
		Entity: entity,
		Value:  value,
		Err:    err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Формирует строго структурированное сообщение, фиксируя целевую сущность и искомый ключ.
func (dnf *DalNotFoundError) Error() string {
	msg := fmt.Sprintf("DAL: %s with value [%v] not found", dnf.Entity, dnf.Value)
	if dnf.Err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов драйвера
		return fmt.Sprintf("%s: %v", msg, dnf.Err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку драйвера СУБД (например, sql.ErrNoRows).
// Требуется для корректной работы функций errors.Is и errors.As во внешних классификаторах транспорта.
func (dnf *DalNotFoundError) Unwrap() error {
	return dnf.Err
}
