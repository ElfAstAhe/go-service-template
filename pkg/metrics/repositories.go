package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Histogram metrics - repository methods latency
var (
	repoDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "repository_op_duration_seconds",
		Help:    "Duration of repository operations",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1, 2.5, 3, 4, 5},
	}, []string{"repository", "method", "status"})
)
