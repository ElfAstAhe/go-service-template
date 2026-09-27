package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RepoDuration — многомерная гистограмма (Histogram) времени задержки (Latency) операций репозиторного слоя.
	// Позволяет SRE и разработчикам вычислять точные перцентили выполнения SQL-запросов (p95, p99),
	// изолировать медленные транзакции и точечно находить узкие места при работе с базами данных (PostgreSQL).
	RepoDuration *prometheus.HistogramVec
)

// InitRepositoryMetrics выполняет потокобезопасную инициализацию и регистрацию метрик репозиторного слоя.
// Метод вызывается один раз на этапе старта приложения внутри инфраструктурного контейнера (InfraContainer).
func InitRepositoryMetrics() {
	// Инициализируем гистограмму с label-тегами для сквозной аналитики:
	// - repository: имя конкретной структуры репозитория (например, "login_attempts_repository")
	// - method: имя вызываемого CRUD-метода (например, "Create", "FindByID")
	// - status: результат операции ("success" или "error") для вычисления Error Rate
	RepoDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "repository_op_duration_seconds",
		Help: "Duration of repository operations in seconds, broken down by repository name, method and execution status.",
		// Бакеты (Buckets) оптимизированы под специфику дискового ввода-вывода (I/O) и транзакций БД:
		// от ультрабыстрого кэша/индексов (1мс) до тяжелых аналитических выборок и сетевых таймаутов (5с).
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1, 2.5, 3, 4, 5},
	}, []string{"repository", "method", "status"})
}
