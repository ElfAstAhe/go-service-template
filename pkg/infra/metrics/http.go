package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal — многомерный счетчик (Counter) общего количества входящих HTTP-запросов.
	// Используется для вычисления RPS (Requests Per Second) и анализа распределения статус-кодов
	// ответов с помощью функции rate() в Grafana.
	HTTPRequestsTotal *prometheus.CounterVec

	// HTTPRequestDuration — многомерная гистограмма (Histogram) времени обработки HTTP-запросов.
	// Позволяет вычислять перцентили задержки (p95, p99) сетевого рантайма и строить графики Latency.
	HTTPRequestDuration *prometheus.HistogramVec
)

// InitHTTPMetrics выполняет потокобезопасную инициализацию и регистрацию системных метрик HTTP-слоя.
// Метод вызывается на этапе старта приложения внутри инфраструктурного контейнера (InfraContainer).
func InitHTTPMetrics() {
	// Инициализируем счетчик запросов с label-тегами для сквозной фильтрации по методам и эндпоинтам
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests broken down by method, path and status code.",
	}, []string{"method", "path", "status"})

	// Инициализируем гистограмму длительности запросов.
	// Бакеты (Buckets) оптимизированы под стандартные тайминги синхронного межсервисного взаимодействия:
	// от ультрабыстрых ответов памяти (5мс) до тяжелых сетевых таймаутов (5с).
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds broken down by method and path.",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"method", "path"})
}
