package errs

import (
	"fmt"
)

// ContainerNotFoundError реализует нативный интерфейс error, специфицируя системные ошибки,
// связанные с поиском или запросом незарегистрированного/отсутствующего DI-контейнера в оркестраторе фреймворка.
//
// Инкапсулирует контекст сбоя: пользовательское текстовое описание (msg)
// и ссылку на исходную нижележащую причину падения рантайма (err).
type ContainerNotFoundError struct {
	msg string // Человекочитаемое описание сути сбоя (например, "target dependency PgContainer was not registered in orchestrator")
	err error  // Ссылка на исходную корневую ошибку (если сбой проброшен из нижележащей системы маппинга DI)
}

// Гарантируем полное соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*ContainerNotFoundError)(nil)

// NewContainerNotFoundError — фабричный конструктор ошибки отсутствия DI-контейнера.
func NewContainerNotFoundError(msg string, err error) *ContainerNotFoundError {
	return &ContainerNotFoundError{msg: msg, err: err}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически собирает текстовую строку, фиксируя факт отсутствия компонента в реестре зависимостей.
func (e *ContainerNotFoundError) Error() string {
	msg := "CNT: not found or not registered"
	if e.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, e.msg)
	}
	if e.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой при поиске контейнера.
// Требуется для корректной работы функций errors.Is и errors.As на этапе стартапа микросервиса.
func (e *ContainerNotFoundError) Unwrap() error {
	return e.err
}
