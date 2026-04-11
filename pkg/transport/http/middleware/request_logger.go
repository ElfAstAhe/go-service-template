package middleware

import (
	"net/http"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type HTTPRequestLogger struct {
	log logger.Logger
}

func NewHTTPRequestLogger(logger logger.Logger) *HTTPRequestLogger {
	return &HTTPRequestLogger{
		log: logger.GetLogger("http_request_logger"),
	}
}

func (hrl *HTTPRequestLogger) Handle(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hrl.log.Debug("HTTPRequestLogger.Handle start")
		defer hrl.log.Debug("HTTPRequestLogger.Handle end")

		wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			span := trace.SpanFromContext(r.Context())
			traceID := ""
			if span.SpanContext().IsValid() {
				traceID = span.SpanContext().TraceID().String()
			}

			// Собираем поля один раз, чтобы не дублировать код
			fields := []any{
				"trace_id", traceID,
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"uri", r.RequestURI,
				"status", wrw.Status(),
				"latency", time.Since(start),
				"bytes", wrw.BytesWritten(),
				"remote_ip", r.RemoteAddr,
			}

			if wrw.Status() >= http.StatusInternalServerError {
				// Если 5xx — это критично
				hrl.log.ErrorW("http request failure", fields...)
			} else if wrw.Status() >= http.StatusBadRequest {
				// Если 4xx — это предупреждение (ошибка клиента)
				hrl.log.WarnW("http request client error", fields...)
			} else {
				// Если всё ОК — пишем в Debug
				hrl.log.DebugW("http request success", fields...)
			}
		}()

		handler.ServeHTTP(wrw, r)
	})
}
