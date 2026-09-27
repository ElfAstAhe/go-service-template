package errs

import (
	"fmt"
)

// BllNotFoundError реализует нативный интерфейс error, специфицируя ошибку логического
// отсутствия запрашиваемого ресурса или доменной модели на уровне бизнес-логики (Domain Layer).
//
// Инкапсулирует расширенный контекст: имя операции (op), название целевой модели (model),
// поисковый идентификатор/ключ (key) и ссылку на исходную причину сбоя (err).
type BllNotFoundError struct {
	op    string // Имя метода UseCase/Сервиса, инициировавшего поиск (например, "GetProfile")
	model string // Наименование отсутствующей доменной сущности (например, "Account", "AuditLog")
	key   string // Строковое представление ключа или ID, по которому производился поиск
	err   error  // Ссылка на исходную корневую ошибку (если сбой проброшен, например, из кэша или DAL)
}

// Гарантируем полное соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*BllNotFoundError)(nil)

// NewBllNotFoundError — фабричный конструктор ошибки отсутствия доменной модели.
func NewBllNotFoundError(op string, model string, key string, err error) *BllNotFoundError {
	return &BllNotFoundError{
		op:    op,
		model: model,
		key:   key,
		err:   err,
	}
}

// Error форматирует и возвращает детализированный строковый паспорт ошибки.
// Формирует строго структурированное сообщение, фиксируя целевую модель, операцию и искомый ключ.
func (nf *BllNotFoundError) Error() string {
	msg := fmt.Sprintf("BLL: model %s not found by op %s with key [%s]", nf.model, nf.op, nf.key)
	if nf.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, nf.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой.
// Обеспечивает корректную работу систем автоматического проброса и маппинга ошибок на транспортном слое.
func (nf *BllNotFoundError) Unwrap() error {
	return nf.err
}
