package middleware

import (
	"net/http"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

// HTTPRequestLogger реализует промежуточный обработчик (Middleware) структурированного,
// контекстно-обогащенного логирования результатов обработки входящих REST/HTTP запросов.
//
// Инкапсулирует под капотом автоматический замер Latency выполнения операции, извлечение
// OTel TraceID, объема записанных в сокет байт и адаптивное логирование уровней Debug/Warn/Error.
type HTTPRequestLogger struct {
	log logger.Logger // Ссылка на центральный структурированный логгер компонента
}

// NewHTTPRequestLogger — фабричный конструктор и инициализатор middleware логирования HTTP-запросов.
func NewHTTPRequestLogger(logger logger.Logger) *HTTPRequestLogger {
	return &HTTPRequestLogger{
		log: logger.GetLogger("http_request_logger"),
	}
}

// Handle встраивает логгер в каскадную цепочку обработки HTTP-запросов (http.Handler Middleware Pattern).
// Перехватывает поток ответа через WrapResponseWriter, вычисляет длительность выполнения и записывает
// в лог-коллектор исчерпывающий структурированный паспорт сетевой операции в формате ключ-значение.
func (hrl *HTTPRequestLogger) Handle(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hrl.log.Debug("HTTPRequestLogger.Handle start")
		defer hrl.log.Debug("HTTPRequestLogger.Handle end")

		// Обертка для перехвата статус-кода и размера ответа стандартной библиотеки
		wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		// Defer-барьер гарантирует запись структурированного лога в коллектор при любом исходе handler-цепочки
		defer func() {
			span := trace.SpanFromContext(r.Context())
			traceID := ""
			if span.SpanContext().IsValid() {
				traceID = span.SpanContext().TraceID().String()
			}

			// Собираем поля один раз, чтобы не дублировать код в условных ветках (DRY паттерн)
			fields := []any{
				"trace_id", traceID,
				"request_id", middleware.GetReqID(r.Context()), // Внимание: завязано на контекстный ключ go-chi
				"method", r.Method,
				"uri", r.RequestURI,
				"status", wrw.Status(),
				"latency", time.Since(start),
				"bytes", wrw.BytesWritten(),
				"remote_ip", r.RemoteAddr,
			}

			// Адаптивная маршрутизация уровня лога на основе итогового HTTP статус-кода ответа
			if wrw.Status() >= http.StatusInternalServerError {
				// Если 5xx — это критично (сбой инфраструктуры или паника UseCase)
				hrl.log.ErrorW("http request failure", fields...)
			} else if wrw.Status() >= http.StatusBadRequest {
				// Если 4xx — это предупреждение (ошибка валидации DTO или инвариантов клиента)
				hrl.log.WarnW("http request client error", fields...)
			} else {
				// Если всё ОК (2xx/3xx) — пишем исключительно в Debug контур во избежание Log Pollution
				hrl.log.DebugW("http request success", fields...)
			}
		}()

		handler.ServeHTTP(wrw, r)
	})
}
