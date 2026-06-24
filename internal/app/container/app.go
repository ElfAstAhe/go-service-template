package container

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

const (
	InstanceApplication       string = "application"
	InstanceApplicationReady  string = "application-ready"
	InstanceApplicationHealth string = "application-health"
	InstanceConfig            string = "config"
	InstanceLogger            string = "logger"
)

type AppContainer struct {
	*container.BaseContainer
}

var _ container.Container = (*AppContainer)(nil)

func NewAppContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *AppContainer {
	return &AppContainer{
		BaseContainer: container.NewBaseContainer(
			container.WithName(AppContainerName),
			container.WithOrchestrator(orchestrator),
			container.WithLogger(log),
		),
	}
}

func (ac *AppContainer) Init(ctx context.Context) error {
	return nil
}

func (ac *AppContainer) Close(ctx context.Context) error {
	return nil
}
