package grpc

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// Option определяет сигнатуру функциональной опции (конфигуратора) для гибкой настройки параметров gRPC-раннера (Fluent API).
type Option func(*Runner)

// WithName задает уникальное строковое имя для gRPC-раннера (используется для изоляции контекстов логирования).
func WithName(name string) Option {
	return func(r *Runner) {
		r.name = name
	}
}

// WithConfig инжектирует предварительно собранную и валидированную структуру конфигурации GRPCConfig.
func WithConfig(conf *config.GRPCConfig) Option {
	return func(r *Runner) {
		r.conf = conf
	}
}

// WithServerProvider инжектирует фабричную функцию (провайдер) для создания экземпляра gRPC сервера.
func WithServerProvider(provider ServerProvider) Option {
	return func(r *Runner) {
		r.serverProvider = provider
	}
}

// WithServiceRegister инжектирует интерфейс/функцию регистрации protobuf-сервисов в gRPC-рантайме.
func WithServiceRegister(register ServiceRegister) Option {
	return func(r *Runner) {
		r.serviceRegister = register
	}
}

// WithServerLauncher инжектирует кастомный механизм запуска/слушателя сетевого сокета gRPC-сервера.
func WithServerLauncher(launcher ServerLauncher) Option {
	return func(r *Runner) {
		r.serverLauncher = launcher
	}
}

// WithLogger инжектирует изолированный структурированный логгер фреймворка для подсистемы наблюдаемости gRPC-раннера.
func WithLogger(name string, log logger.Logger) Option {
	return func(r *Runner) {
		r.log = log.GetLogger(name)
	}
}

// WithShutdownTimeout гранулярно переопределяет временной потолок на мягкое закрытие (Graceful Shutdown) gRPC-сервера.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(r *Runner) {
		r.conf.ShutdownTimeout = timeout
	}
}

// WithAppEnv инжектирует текущую среду окружения рантайма (dev/prod/test) для адаптации политик безопасности gRPC.
func WithAppEnv(env config.AppEnv) Option {
	return func(r *Runner) {
		r.env = env
	}
}
