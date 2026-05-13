package container

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
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
) *AppContainer {
	return &AppContainer{
		BaseContainer: container.NewBaseContainer(AppContainerName, orchestrator),
	}
}

func (ac *AppContainer) Init(ctx context.Context) error {
	return nil
}
