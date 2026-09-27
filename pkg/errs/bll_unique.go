package errs

import (
	"fmt"
)

// BllUniqueError реализует нативный интерфейс error, специфицируя ошибку нарушения
// уникальности данных (коллизия дублирования) на уровне бизнес-логики (Domain Layer).
//
// Инкапсулирует контекст падения: имя доменной операции (op), название целевой модели (model),
// имя или значение дублирующегося ключа (key) и ссылку на исходную нижележащую причину сбоя (err).
type BllUniqueError struct {
	op    string // Имя метода UseCase/Сервиса, где обнаружено дублирование данных (например, "Register")
	model string // Наименование доменной сущности, давшей конфликт уникальности (например, "User", "Client")
	key   string // Название поля или значение, вызвавшее коллизию (например, "email", "inn")
	err   error  // Ссылка на исходную корневую ошибку (например, ошибку уникальности DAL-слоя из PostgreSQL)
}

// Гарантируем полное соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*BllUniqueError)(nil)

// NewBllUniqueError — фабричный конструктор ошибки дублирования/уникальности модели.
func NewBllUniqueError(op string, model string, key string, err error) *BllUniqueError {
	return &BllUniqueError{
		op:    op,
		model: model,
		key:   key,
		err:   err,
	}
}

// Error форматирует и возвращает финальное строковое представление ошибки.
// Последовательно склеивает название модели, имя операции, конфликтующий ключ и стек вложенных ошибок.
func (u *BllUniqueError) Error() string {
	msg := fmt.Sprintf("BLL: %s model already exists, op %s", u.model, u.op)
	if u.key != "" {
		msg = fmt.Sprintf("%s key [%s]", msg, u.key)
	}
	if u.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, u.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой.
// Обеспечивает корректную работу систем автоматического проброса и маппинга ошибок на транспортном слое.
func (u *BllUniqueError) Unwrap() error {
	return u.err
}
