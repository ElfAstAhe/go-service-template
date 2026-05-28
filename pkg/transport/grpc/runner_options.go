package grpc

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

func WithConfig(conf *config.GRPCConfig) Option {
	return func(r *Runner) {
		r.conf = conf
	}
}

func WithServerProvider(provider ServerProvider) Option {
	return func(r *Runner) {
		r.serverProvider = provider
	}
}

func WithServiceRegister(register ServiceRegister) Option {
	return func(r *Runner) {
		r.serviceRegister = register
	}
}

func WithServerLauncher(launcher ServerLauncher) Option {
	return func(r *Runner) {
		r.serverLauncher = launcher
	}
}

func WithLogger(name string, log logger.Logger) Option {
	return func(r *Runner) {
		r.log = log.GetLogger(name)
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(r *Runner) {
		r.conf.ShutdownTimeout = timeout
	}
}

func WithAppEnv(env config.AppEnv) Option {
	return func(r *Runner) {
		r.env = env
	}
}
