package worker

import (
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
)

// Janitor представляет собой фоновый строго типизированный воркер-планировщик (Scheduler Worker).
//
// Инкапсулирует работу с BaseScheduler фреймворка, периодически запуская такт проактивной
// очистки просроченных по TTL записей (Garbage Collection) в оперативной памяти кэш-системы.
type Janitor struct {
	*worker.BaseScheduler // Встраиваем базовый планировщик для управления жизненным циклом и тикерами
}

// NewJanitor — фабричный конструктор фонового очистителя кэша.
// Прозрачно связывает метод CacheJanitor кэш-менеджера со встроенным циклом планировщика задач.
func NewJanitor[K comparable, V any](
	name string,
	conf *worker.BaseSchedulerConfig,
	c cache.Cache[K, V],
	log logger.Logger,
) *Janitor {
	return &Janitor{
		BaseScheduler: worker.NewBaseScheduler(
			name,
			c.CacheJanitor, // Инжектируем метод свипинга кэша в качестве целевой фоновой задачи
			conf,
			log,
		),
	}
}
