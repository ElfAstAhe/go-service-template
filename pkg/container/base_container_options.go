package container

import (
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type Option func(*Options)

type Options struct {
	Name         string
	Orchestrator Orchestrator
	Logger       logger.Logger
}

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithOrchestrator(orchestrator Orchestrator) Option {
	return func(o *Options) {
		o.Orchestrator = orchestrator
	}
}

func WithLogger(logger logger.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}
