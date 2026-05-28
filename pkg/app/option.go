package app

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type Option func(*BaseApplication)

func WithLogger(logger logger.Logger) Option {
	return func(app *BaseApplication) {
		app.logger = logger
	}
}

func WithOrchestrator(orchestrator container.Orchestrator) Option {
	return func(app *BaseApplication) {
		app.orchestrator = orchestrator
	}
}

func WithConfig(appConf *config.AppConfig) Option {
	return func(app *BaseApplication) {
		app.conf = appConf
	}
}

func WithStopTimeout(timeout time.Duration) Option {
	return func(app *BaseApplication) {
		app.conf.StopTimeout = timeout
	}
}

func WithCloseTimeout(timeout time.Duration) Option {
	return func(app *BaseApplication) {
		app.conf.CloseTimeout = timeout
	}
}

func WithInitTimeout(timeout time.Duration) Option {
	return func(app *BaseApplication) {
		app.conf.InitTimeout = timeout
	}
}
