package errs

import (
	"fmt"
)

// ContainerError реализует нативный интерфейс error, специфицируя системные ошибки сборки,
// инициализации и разрешения зависимостей внутри DI-контейнеров (Dependency Injection Container Error).
//
// Инкапсулирует контекст сбоя: имя контейнера (name), пользовательское текстовое описание (msg)
// и ссылку на исходную нижележащую причину падения рантайма (err).
type ContainerError struct {
	name string // Строковое имя упавшего DI-контейнера (например, "PgContainer", "WorkerContainer")
	msg  string // Человекочитаемое описание сути сбоя инициализации (например, "failed to resolve consumer dependency")
	err  error  // Ссылка на исходную корневую ошибку (например, системную ошибку сетевого драйвера базы данных)
}

// Гарантируем полное соответствие встроенному интерфейсу error на этапе компиляции
var _ error = (*ContainerError)(nil)

// NewContainerError — фабричный конструктор ошибки инициализации DI-контейнера.
func NewContainerError(name string, msg string, err error) *ContainerError {
	return &ContainerError{
		name: name,
		msg:  msg,
		err:  err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически собирает текстовую строку, четко изолируя сбойный контейнер и прикрепляя корневую ошибку.
func (ce *ContainerError) Error() string {
	msg := "CNT: common container error"
	if ce.name != "" {
		msg = fmt.Sprintf("%s, container [%s]", msg, ce.name)
	} else {
		// Fallback предохранитель на случай, если имя контейнера передано пустым
		msg = fmt.Sprintf("%s, container [unknown]", msg)
	}
	if ce.msg != "" {
		msg = fmt.Sprintf("%s with message %s", msg, ce.msg)
	}
	if ce.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, ce.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой при сборке зависимостей.
// Требуется для корректной работы функций errors.Is и errors.As на этапе стартапа микросервиса.
func (ce *ContainerError) Unwrap() error {
	return ce.err
}
