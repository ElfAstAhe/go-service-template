package errs

import (
	"fmt"
)

// BllForbiddenError реализует нативный интерфейс error, специфицируя ошибку нарушения прав
// или отсутствия необходимых ролей/скоупов (RBAC/ABAC) на уровне бизнес-логики (Domain Layer).
//
// Инкапсулирует контекст падения: имя доменной операции (op), пользовательское текстовое сообщение (msg)
// и ссылку на исходную нижележащую причину сбоя (err).
type BllForbiddenError struct {
	op  string // Имя защищенного метода UseCase, где произошел отказ в доступе (например, "DeleteUser")
	msg string // Прикладное пояснение причины отказа (например, "only administrator can perform this action")
	err error  // Ссылка на исходную корневую ошибку (если сбой произошел, например, внутри внешней подсистемы IAM)
}

// Гарантируем полное соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*BllForbiddenError)(nil)

// NewBllForbiddenError — фабричный конструктор ошибки нехватки прав доступа.
func NewBllForbiddenError(op string, msg string, err error) *BllForbiddenError {
	return &BllForbiddenError{op, msg, err}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Последовательно склеивает контекст операции, текстовое сообщение и стек вложенных ошибок.
func (bfe *BllForbiddenError) Error() string {
	msg := "BLL: forbidden"
	if bfe.op != "" {
		msg = fmt.Sprintf("%s at operation %s", msg, bfe.op)
	}
	if bfe.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, bfe.msg)
	}
	if bfe.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя
		msg = fmt.Sprintf("%s: %v", msg, bfe.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой.
// Необходим для корректной работы систем размотки стека ошибок (errors.Is / errors.As) на транспортном слое.
func (bfe *BllForbiddenError) Unwrap() error {
	return bfe.err
}
