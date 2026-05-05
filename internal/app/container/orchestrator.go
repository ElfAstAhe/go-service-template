package container

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
)

type Orchestrator struct {
	*container.BaseOrchestrator
	conf *config.Config
}

var _ container.Orchestrator = (*Orchestrator)(nil)

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{}
}

func (o *Orchestrator) Init(ctx context.Context) error {
	return nil
}
