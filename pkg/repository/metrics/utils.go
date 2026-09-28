package metrics

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/metrics"
)

// ObserveRepositoryOp — утилитарный хелпер, рассчитывающий статус выполнения операции
// и атомарно пушащий тайминги задержки (Latency) в гистограмму RepoDuration.
//
// Вызывается автоматически внутри отложенных вызовов (defer) дженерик-декораторов репозиториев.
func ObserveRepositoryOp(repository, method string, err error, startTime time.Time) {
	// По умолчанию выставляем статус успешного завершения операции
	status := metrics.StatusSuccess
	if err != nil {
		// При наличии любой ошибки (сетевой или логической) фиксируем сбой
		status = metrics.StatusFail
	}

	// Рассчитываем дельту времени выполнения операции в секундах (float64) и логируем в Prometheus
	metrics.RepoDuration.WithLabelValues(repository, method, status).Observe(time.Since(startTime).Seconds())
}
