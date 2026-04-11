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

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Обертка для перехвата статус-кода и размера ответа
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			// 1. Получаем шаблон пути: /api/test/{id} вместо /api/test/123
			path := chi.RouteContext(r.Context()).RoutePattern()
			if path == "" {
				path = "unknown"
			}

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(ww.Status())

			// 2. Вытаскиваем TraceID для Exemplar
			var exemplar prometheus.Labels
			if span := trace.SpanContextFromContext(r.Context()); span.IsSampled() {
				exemplar = prometheus.Labels{"traceID": span.TraceID().String()}
			}

			// 3. Записываем Counter
			metrics.HttpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()

			// 4. Записываем Histogram с Exemplar (Senior-way)
			observer := metrics.HttpRequestDuration.WithLabelValues(r.Method, path)
			if exemplar != nil {
				if ex, ok := observer.(prometheus.ExemplarObserver); ok {
					ex.ObserveWithExemplar(duration, exemplar)
				} else {
					observer.Observe(duration)
				}
			} else {
				observer.Observe(duration)
			}
		}()

		next.ServeHTTP(ww, r)
	})
}
