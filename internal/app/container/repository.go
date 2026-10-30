package container

import (
	"context"
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

const (
	InstanceTestRepo  string = "testRepo"
	InstanceTest2Repo string = "test2-repository"

	InstanceTestMetrics  string = "test-metrics-repo"
	InstanceTest2Metrics string = "test2-metrics-repo"

	InstanceTestTrace  string = "test-trace-repo"
	InstanceTest2Trace string = "test2-trace-repo"
)

type RepositoryContainer struct {
	*container.BaseLazyContainer
}

var _ container.Container = (*RepositoryContainer)(nil)
var _ container.LazyContainer = (*RepositoryContainer)(nil)

func NewRepositoryContainer(
	orchestrator container.Orchestrator,
	log logger.Logger,
) *RepositoryContainer {
	return &RepositoryContainer{
		BaseLazyContainer: container.NewBaseLazyContainer(
			container.WithLazyName(RepositoryContainerName),
			container.WithLazyOrchestrator(orchestrator),
			container.WithLazyLogger(log),
		),
	}
}

//goland:noinspection DuplicatedCode
func (rc *RepositoryContainer) Init(ctx context.Context) error {
	err := errors.Join(
		rc.RegisterProvider(InstanceTestRepo, rc.providerTestRepository),
		rc.RegisterProvider(InstanceTest2Repo, rc.providerTest2Repository),
		rc.RegisterProvider(InstanceTestMetrics, rc.providerTestMetrics),
		rc.RegisterProvider(InstanceTest2Metrics, rc.providerTest2Metrics),
		rc.RegisterProvider(InstanceTestTrace, rc.providerTestTrace),
		rc.RegisterProvider(InstanceTest2Trace, rc.providerTest2Trace),
	)
	if err != nil {
		return errs.NewContainerError(rc.GetName(), "container init: register providers failed", err)
	}

	return nil
}
