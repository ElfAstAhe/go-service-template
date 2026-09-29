package container

import (
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// LazyOption определяет сигнатуру функциональной опции (конфигуратора) для ленивых контейнеров.
type LazyOption func(*LazyOptions)

// LazyOptions инкапсулирует параметры конфигурации подсистемы отложенной инициализации зависимостей.
// Переиспользует базовые свойства структуры Options за счет композиции (встраивания указателя).
type LazyOptions struct {
	*Options // Встроенный указатель на базовый набор сквозных зависимостей фреймворка
}

// WithLazyName задает уникальное строковое имя для инициализируемого ленивого контейнера.
// Безопасно аллоцирует базовые опции, предотвращая паники времени выполнения (Null-Pointer Protection).
func WithLazyName(name string) LazyOption {
	return func(o *LazyOptions) {
		if o.Options == nil {
			o.Options = &Options{}
		}
		o.Name = name
	}
}

// WithLazyOrchestrator принудительно связывает ленивый контейнер с глобальным DI-оркестратором.
func WithLazyOrchestrator(orchestrator Orchestrator) LazyOption {
	return func(o *LazyOptions) {
		if o.Options == nil {
			o.Options = &Options{}
		}
		o.Orchestrator = orchestrator
	}
}

// WithLazyLogger инжектирует структурированный логгер фреймворка в ленивый контейнер.
func WithLazyLogger(logger logger.Logger) LazyOption {
	return func(o *LazyOptions) {
		if o.Options == nil {
			o.Options = &Options{}
		}
		o.Logger = logger
	}
}
