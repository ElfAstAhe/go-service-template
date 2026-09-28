package http

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// Option определяет сигнатуру функциональной опции (конфигуратора) для гибкой настройки параметров HTTP-раннера (Fluent API).
type Option func(*Runner)

// WithName задает уникальное строковое имя для HTTP-раннера (используется для изоляции контекстов логирования и метрик).
func WithName(name string) Option {
	return func(r *Runner) {
		r.name = name
	}
}

// WithRouter инжектирует в раннер реализацию интерфейса Router, предоставляющую готовый сетевой http.Handler.
func WithRouter(router Router) Option {
	return func(r *Runner) {
		r.router = router
	}
}

// WithConfig инжектирует предварительно собранную и валидированную структуру конфигурации HTTPConfig веб-сервера.
func WithConfig(conf *config.HTTPConfig) Option {
	return func(r *Runner) {
		r.conf = conf
	}
}

// WithServerProvider инжектирует фабричную функцию (провайдер) для кастомной сборки экземпляра веб-сервера.
func WithServerProvider(provider ServerProvider) Option {
	return func(r *Runner) {
		r.serverProvider = provider
	}
}

// WithServerLauncher инжектирует кастомную стратегию запуска сетевого слушателя (например, для мокирования сокетов).
func WithServerLauncher(launcher ServerLauncher) Option {
	return func(r *Runner) {
		r.serverLauncher = launcher
	}
}

// WithShutdownTimeout гранулярно переопределяет временной потолок на мягкое закрытие (Graceful Shutdown) HTTP-сервера.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(r *Runner) {
		r.conf.ShutdownTimeout = timeout
	}
}

// WithLogger инжектирует изолированный структурированный логгер фреймворка для подсистемы наблюдаемости HTTP-раннера.
func WithLogger(name string, log logger.Logger) Option {
	return func(r *Runner) {
		r.log = log.GetLogger(name)
	}
}
