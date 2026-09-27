package errs

import (
	"fmt"
)

// ConfigError реализует нативный интерфейс error, специфицируя системные ошибки
// парсинга, валидации и загрузки параметров конфигурации приложения (Configuration Layer Error).
//
// Инкапсулирует контекст сбоя: пользовательское текстовое описание проблемы (msg)
// и ссылку на исходную нижележащую причину падения рантайма (err).
type ConfigError struct {
	msg string // Человекочитаемое описание сути сбоя парсинга (например, "database port must be a valid integer")
	err error  // Ссылка на исходную корневую ошибку (например, ошибку разбора структуры из библиотеки Viper)
}

// Гарантируем соответствие интерфейсу error на этапе компиляции
var _ error = (*ConfigError)(nil)

// NewConfigError — фабричный конструктор ошибки конфигурационного слоя.
func NewConfigError(msg string, err error) *ConfigError {
	return &ConfigError{
		msg: msg,
		err: err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически собирает текстовую строку, четко изолируя сбои подсистемы парсинга параметров.
func (e *ConfigError) Error() string {
	msg := "CFG: error"
	if e.msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.msg)
	}
	if e.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, e.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой при загрузке параметров.
// Требуется для корректной работы функций errors.Is и errors.As на этапе инициализации микросервиса.
func (e *ConfigError) Unwrap() error {
	return e.err
}
