package container

import (
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// Option определяет сигнатуру функциональной опции (конфигуратора) для настройки параметров контейнера (Fluent API).
type Option func(*Options)

// Options инкапсулирует конфигурационные метаданные и общие сквозные зависимости,
// необходимые для инициализации, сборки и логирования жизненного цикла DI-контейнера.
type Options struct {
	Name         string        // Уникальное текстовое имя контейнера (например, "KafkaContainer")
	Orchestrator Orchestrator  // Ссылка на центральный IoC-оркестратор для интеграции в общий граф зависимостей
	Logger       logger.Logger // Экземпляр логгера для вывода отладочных шагов сборки и валидации компонентов
}

// WithName задает уникальное строковое имя для инициализируемого контейнера.
func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// WithOrchestrator принудительно привязывает контейнер к целевому центральному DI-оркестратору.
func WithOrchestrator(orchestrator Orchestrator) Option {
	return func(o *Options) {
		o.Orchestrator = orchestrator
	}
}

// WithLogger инжектирует структурированный логгер фреймворка для подсистемы наблюдаемости контейнера.
func WithLogger(logger logger.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}
