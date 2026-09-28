package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

// MetricsMiddleware осуществляет сквозной перехват REST/HTTP запросов для сбора
// телеметрии объема трафика и гистограмм задержек (Latency) в формате Prometheus.
//
// 📊 Реализованные паттерны наблюдаемости (Observability):
//  1. Агрегация путей по маскам роутера (Route Pattern) для предотвращения High Cardinality взрыва памяти.
//  2. Инжекция OpenTelemetry TraceID прямо в гистограммы Prometheus (Паттерн OpenMetrics Exemplars)
//     для бесшовной сквозной корреляции графиков и распределенных трейсов в UI Grafana.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Обертка для перехвата статус-кода и размера ответа стандартной библиотеки
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		// Defer-барьер гарантирует запись метрик даже в случае аварийного завершения запроса
		defer func() {
			// 1. Получаем шаблон пути: /api/test/{id} вместо /api/test/123
			path := chi.RouteContext(r.Context()).RoutePattern()
			if path == "" {
				path = "unknown"
			}

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(ww.Status())

			// 2. Вытаскиваем TraceID для Exemplar из контекста OpenTelemetry
			var exemplar prometheus.Labels
			if span := trace.SpanContextFromContext(r.Context()); span.IsSampled() {
				exemplar = prometheus.Labels{"traceID": span.TraceID().String()}
			}

			// 3. Записываем Counter общего количества входящих запросов по методам и путям
			metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()

			// 4. Записываем Histogram с Exemplar (Senior-way) для детального анализа перцентилей Latency
			observer := metrics.HTTPRequestDuration.WithLabelValues(r.Method, path)
			if exemplar != nil {
				// Динамически проверяем поддержку интерфейса ExemplarObserver текущим драйвером
				if ex, ok := observer.(prometheus.ExemplarObserver); ok {
					ex.ObserveWithExemplar(duration, exemplar)
				} else {
					observer.Observe(duration)
				}
			} else {
				observer.Observe(duration)
			}
		}()

		// Передаем управление следующему обработчику в каскаде HTTP-декораторов
		next.ServeHTTP(ww, r)
	})
}
