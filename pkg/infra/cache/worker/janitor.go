package worker

import (
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
)

type Janitor struct {
	*worker.BaseScheduler
}

func NewJanitor[K comparable, V any](
	name string,
	conf *worker.BaseSchedulerConfig,
	c cache.Cache[K, V],
	log logger.Logger,
) *Janitor {
	return &Janitor{
		BaseScheduler: worker.NewBaseScheduler(
			name,
			c.CacheJanitor,
			conf,
			log,
		),
	}
}
