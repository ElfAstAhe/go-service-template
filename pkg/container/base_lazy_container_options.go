package container

import (
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type LazyOption func(*LazyOptions)

type LazyOptions struct {
	*Options
}

func WithLazyName(name string) LazyOption {
	return func(o *LazyOptions) {
		if o.Options == nil {
			o.Options = &Options{}
		}
		o.Options.Name = name
	}
}

func WithLazyOrchestrator(orchestrator Orchestrator) LazyOption {
	return func(o *LazyOptions) {
		if o.Options == nil {
			o.Options = &Options{}
		}
		o.Options.Orchestrator = orchestrator
	}
}

func WithLazyLogger(logger logger.Logger) LazyOption {
	return func(o *LazyOptions) {
		if o.Options == nil {
			o.Options = &Options{}
		}
		o.Options.Logger = logger
	}
}
