package middleware

import (
	"fmt"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
)

// Константы стандартных HTTP-заголовков для сквозного отслеживания цепочек распределенных вызовов.
const (
	HeaderXRequestID     string = "X-Request-ID"     // Классический сквозной идентификатор атомарного запроса
	HeaderXCorrelationID string = "X-Correlation-ID" // Сквозной идентификатор комплексной бизнес-цепочки (Correlation)
	HeaderRequestID      string = "Request-ID"       // Дублирующий маркер для обратной совместимости систем
)

// RequestIDExtractor инкапсулирует в себе перечень отслеживаемых HTTP-заголовков
// для извлечения или автоматической генерации сквозного идентификатора запроса (Request-ID Extraction Core).
type RequestIDExtractor struct {
	headers []string // Приоритетный срез имен заголовков для последовательного сканирования
}

// NewRequestIDExtractor — фабричный конструктор экстрактора с произвольным набором отслеживаемых заголовков.
func NewRequestIDExtractor(headers ...string) *RequestIDExtractor {
	return &RequestIDExtractor{
		headers: headers,
	}
}

// NewDefaultRequestIDExtractor собирает экстрактор, наполненный базовыми дефолтными заголовками трассировки фреймворка.
func NewDefaultRequestIDExtractor() *RequestIDExtractor {
	return NewRequestIDExtractor(
		HeaderXRequestID,
		HeaderXCorrelationID,
		HeaderRequestID,
	)
}

// Handler встраивает экстрактор в каскадную цепочку обработки HTTP-запросов (http.Handler Middleware Pattern).
// Последовательно сканирует заголовки, осуществляет генерацию фолбека при их отсутствии, обогащает
// context.Context идентификатором и передает управление дальше по стеку декораторов.
func (re *RequestIDExtractor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Fast-Path: если список заголовков пуст — мгновенно передаем управление дальше без вычислений
		if len(re.headers) == 0 {
			next.ServeHTTP(rw, r)
			return
		}

		var requestID string
		// Последовательно сканируем заголовки до первого совпадения (First-Match стратегия)
		for _, header := range re.headers {
			requestID = r.Header.Get(header)
			if requestID != "" {
				break
			}
		}

		// Фолбек: если маркер не передан извне — генерируем детерминированный ID на базе префикса и атомарного счетчика
		if requestID == "" {
			requestID = fmt.Sprintf("%s-%07d", transport.GetPrefix(), transport.NextReqID())
		}

		// Декорируем запрос, упаковав RequestID в context.Context горутины
		next.ServeHTTP(rw, r.WithContext(transport.WithRequestID(r.Context(), requestID)))
	})
}
