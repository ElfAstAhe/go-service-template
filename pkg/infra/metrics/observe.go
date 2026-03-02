package metrics

import (
	"time"
)

func ObserveRepositoryOp(repository, method string, err error, startTime time.Time) {
	status := StatusSuccess
	if err != nil {
		status = StatusFail
	}

	repoDuration.WithLabelValues(repository, method, status).Observe(time.Since(startTime).Seconds())
}
