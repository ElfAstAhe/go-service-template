package errs

import (
	"fmt"
)

// ConfigValidateError реализует нативный интерфейс error, специфицируя системные ошибки
// семантической проверки и валидации параметров конфигурации на этапе запуска (Bootstrap Phase).
//
// Инкапсулирует структурированный контекст сбоя: логическую группу настроек (group), имя конкретного
// ошибочного параметра (item), прикладное описание сути нарушения (msg) и ссылку на исходную причину (err).
type ConfigValidateError struct {
	group string // Логический блок или секция конфигурации (например, "postgres", "telemetry")
	item  string // Имя конкретного невалидного ключа/переменной (например, "port", "sample_rate")
	msg   string // Человекочитаемое описание сути нарушения инварианта параметров
	err   error  // Ссылка на нижележащую ошибку рантайма (например, ошибку парсинга типов)
}

// Гарантируем строгое соответствие интерфейсу error на этапе компиляции
var _ error = (*ConfigValidateError)(nil)

// NewConfigValidateError — фабричный конструктор ошибки валидации параметров конфигурации.
func NewConfigValidateError(group string, item string, msg string, err error) *ConfigValidateError {
	return &ConfigValidateError{
		group: group,
		item:  item,
		msg:   msg,
		err:   err,
	}
}

// Error форматирует и возвращает развернутый строковый паспорт ошибки.
// Динамически выстраивает цепочку метаданных от верхнеуровневой группы до конкретного элемента и текста ошибки.
func (cv *ConfigValidateError) Error() string {
	msg := "CFG: validate error"
	if cv.group != "" {
		msg = fmt.Sprintf("%s: group %s", msg, cv.group)
	}
	if cv.item != "" {
		msg = fmt.Sprintf("%s: item %s", msg, cv.item)
	}
	if cv.msg != "" {
		msg = fmt.Sprintf("%s: msg %s", msg, cv.msg)
	}
	if cv.err != nil {
		// Обогащаем текстовый вывод описанием нижележащего сбоя из стека вызовов
		msg = fmt.Sprintf("%s: %v", msg, cv.err)
	}

	return msg
}

// Unwrap возвращает исходную ошибку, инициировавшую сбой при семантической проверке параметров.
// Требуется для корректной работы функций errors.Is и errors.As на этапе инициализации микросервиса.
func (cv *ConfigValidateError) Unwrap() error {
	return cv.err
}
