package middleware

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
)

// Константы стандартных и проприетарных HTTP-заголовков для ведения распределенного трейсинга (Distributed Tracing).
const (
	HeaderXCloudTraceContext string = "X-Cloud-Trace-Context" // Спецификация трассировки Google Cloud Platform (GCP)
	HeaderTraceParent        string = "Traceparent"           // Каноничный международный стандарт консорциума W3C Trace Context
	HeaderXTraceID           string = "X-Trace-ID"            // Распространенный альтернативный enterprise-заголовок трассировки
	HeaderTraceID            string = "Trace-ID"              // Упрощенный плоский текстовый маркер идентификатора следа
)

// TraceIDExtractor инкапсулирует в себе перечень отслеживаемых HTTP-заголовков
// для извлечения сквозного идентификатора трассировки из входящего сетевого трафика.
type TraceIDExtractor struct {
	headers []string // Приоритетный срез имен заголовков для последовательного сканирования
}

// NewTraceIDExtractor — фабричный конструктор экстрактора с произвольным набором отслеживаемых заголовков трассировки.
func NewTraceIDExtractor(headers ...string) *TraceIDExtractor {
	return &TraceIDExtractor{
		headers: headers,
	}
}

// NewDefaultTraceIDExtractor собирает экстрактор, наполненный базовыми дефолтными заголовками трассировки фреймворка.
func NewDefaultTraceIDExtractor() *TraceIDExtractor {
	return NewTraceIDExtractor(
		HeaderXCloudTraceContext,
		HeaderTraceParent,
		HeaderXTraceID,
		HeaderTraceID,
	)
}

// Handler встраивает экстрактор в каскадную цепочку обработки HTTP-запросов (http.Handler Middleware Pattern).
// Последовательно сканирует заголовки, осуществляет извлечение TraceID (First-Match стратегия),
// обогащает context.Context идентификатором и передает управление дальше по стеку декораторов.
func (te *TraceIDExtractor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Fast-Path: если список заголовков пуст — мгновенно передаем управление дальше без вычислений
		if len(te.headers) == 0 {
			next.ServeHTTP(rw, r)
			return
		}

		var traceID string
		// Последовательно сканируем заголовки до первого совпадения (First-Match стратегия)
		for _, header := range te.headers {
			traceID = r.Header.Get(header)
			if traceID != "" {
				break
			}
		}

		// Декорируем запрос, упаковав TraceID в context.Context горутины
		next.ServeHTTP(rw, r.WithContext(transport.WithTraceID(r.Context(), traceID)))
	})
}
