package errs

import (
	"fmt"
)

// ContainerValidateError реализует нативный интерфейс error, специфицируя системные ошибки
// валидации параметров конфигурации внутри DI-контейнеров на этапе стартапа (Bootstrap Phase).
//
// Инкапсулирует расширенный контекст: имя контейнера (name), операцию/фазу (op),
// пользовательское текстовое сообщение (msg) и ссылку на исходную причину сбоя (err).
type ContainerValidateError struct {
	name string // Имя валидируемого DI-контейнера (например, "KafkaContainer", "CacheContainer")
	op   string // Конкретная фаза или метод валидации (например, "ValidateConfig", "BuildStorage")
	msg  string // Человекочитаемое прикладное описание сути нарушения инварианта параметров
	err  error  // Ссылка на исходную корневую ошибку (если сбой проброшен из внешних систем проверки)
}

// Гарантируем строгое соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*ContainerValidateError)(nil)

// NewContainerValidateError — фабричный конструктор ошибки валидации параметров DI-контейнера.
func NewContainerValidateError(name, op, msg string, err error) *ContainerValidateError {
	return &ContainerValidateError{
		name: name,
		op:   op,
		msg:  msg,
		err:  err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Последовательно склеивает имя контейнера, операцию, текстовое сообщение и стек вложенных ошибок.
func (cve *ContainerValidateError) Error() string {
	msg := "CNT: validate error"
	if cve.name != "" {
		msg = fmt.Sprintf("%s container [%s]", msg, cve.name)
	}
	if cve.op != "" {
		msg = fmt.Sprintf("%s operation [%s]", msg, cve.op)
	}
	if cve.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, cve.msg)
	}
	if cve.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, cve.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой при валидации контейнера.
// Требуется для корректной работы функций errors.Is и errors.As на этапе стартапа микросервиса.
func (cve *ContainerValidateError) Unwrap() error {
	return cve.err
}
