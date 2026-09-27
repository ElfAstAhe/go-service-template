package app

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// Option определяет сигнатуру функциональной опции (конфигуратора) для гибкой настройки параметров приложения (Fluent API).
type Option func(*BaseApplication)

// WithLogger инжектирует центральный структурированный логгер верхнего уровня для подсистемы наблюдаемости рантайма.
func WithLogger(logger logger.Logger) Option {
	return func(app *BaseApplication) {
		app.logger = logger
	}
}

// WithOrchestrator принудительно привязывает приложение к центральному IoC/DI оркестратору графа зависимостей.
func WithOrchestrator(orchestrator container.Orchestrator) Option {
	return func(app *BaseApplication) {
		app.orchestrator = orchestrator
	}
}

// WithConfig инжектирует предварительно собранную и валидированную структуру глобальной конфигурации AppConfig.
func WithConfig(appConf *config.AppConfig) Option {
	return func(app *BaseApplication) {
		app.conf = appConf
	}
}

// WithStopTimeout гранулярно переопределяет временной потолок на мягкую остановку серверов и воркеров брокеров очередей.
func WithStopTimeout(timeout time.Duration) Option {
	return func(app *BaseApplication) {
		app.conf.StopTimeout = timeout
	}
}

// WithCloseTimeout гранулярно переопределяет временной потолок на каскадное закрытие и деаллокацию пулов памяти/сокетов СУБД.
func WithCloseTimeout(timeout time.Duration) Option {
	return func(app *BaseApplication) {
		app.conf.CloseTimeout = timeout
	}
}

// WithInitTimeout гранулярно переопределяет временной потолок на холодную сборку и инициализацию графа IoC/DI-контейнеров.
func WithInitTimeout(timeout time.Duration) Option {
	return func(app *BaseApplication) {
		app.conf.InitTimeout = timeout
	}
}
