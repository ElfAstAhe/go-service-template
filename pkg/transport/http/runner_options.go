package http

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type Option func(*Runner)

func WithName(name string) Option {
	return func(r *Runner) {
		r.name = name
	}
}
func WithRouter(router Router) Option {
	return func(r *Runner) {
		r.router = router
	}
}

func WithConfig(conf *config.HTTPConfig) Option {
	return func(r *Runner) {
		r.conf = conf
	}
}

func WithServerProvider(provider ServerProvider) Option {
	return func(r *Runner) {
		r.serverProvider = provider
	}
}

func WithServerLauncher(launcher ServerLauncher) Option {
	return func(r *Runner) {
		r.serverLauncher = launcher
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(r *Runner) {
		r.conf.ShutdownTimeout = timeout
	}
}

func WithLogger(name string, log logger.Logger) Option {
	return func(r *Runner) {
		r.log = log.GetLogger(name)
	}
}
