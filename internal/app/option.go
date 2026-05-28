package app

import (
	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type Option func(*Application)

func WithConfig(conf *config.Config) Option {
	return func(app *Application) {
		app.conf = conf
	}
}

func WithLogger(log logger.Logger) Option {
	return func(app *Application) {
		app.log = log
	}
}
