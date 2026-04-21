package app

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
)

type Orchestrator struct {
	*container.BaseOrchestrator
	conf *config.Config
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{}
}

func (o *Orchestrator) Init(ctx context.Context) error {}
