package interceptors

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Константы gRPC метаданных для распределенного сквозного трассирования (Distributed Tracing).
const (
	MDXCloudTraceContext string = "x-cloud-trace-context" // Формат трассировки провайдеров Google Cloud Platform (GCP)
	MDTraceParent        string = "traceparent"           // Каноничный международный стандарт консорциума W3C Trace Context
	MDXTraceID           string = "x-trace-id"            // Распространенный кастомный enterprise-заголовок трассировки
	MDTraceID            string = "trace-id"              // Плоский альтернативный маркер идентификатора спана
)

// TraceIDExtractorUSInterceptor возвращает унарный интерцептор (Unary Server Interceptor),
// отвечающий за последовательный поиск и извлечение сквозного идентификатора трассировки TraceID
// из входящих gRPC метаданных HTTP/2 с последующей фиксацией значения в контексте горутины.
func TraceIDExtractorUSInterceptor(headers ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var traceID string
		// Последовательно сканируем переданный приоритетный срез метаданных (First-Match)
		for _, header := range headers {
			vals := metadata.ValueFromIncomingContext(ctx, header)
			if len(vals) > 0 {
				traceID = vals[0]
				break // Извлекаем первый валидный непустой маркер трассировки и прерываем цикл
			}
		}

		// Передаем управление дальше по цепочке интерцепторов, упаковав TraceID в context.Context
		return handler(transport.WithTraceID(ctx, traceID), req)
	}
}

// TraceIDExtractorSSInterceptor возвращает стриминговый интерцептор (Stream Server Interceptor),
// гарантирующий непрерывное сквозное трассирование, связывание и ведение логов для долгоживущих gRPC-потоков (Streams).
func TraceIDExtractorSSInterceptor(headers ...string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var traceID string
		// Сканируем входящие метаданные сетевого HTTP/2 потока данных
		for _, header := range headers {
			vals := metadata.ValueFromIncomingContext(ss.Context(), header)
			if len(vals) > 0 {
				traceID = vals[0]
				break
			}
		}

		// Оборачиваем исходный ServerStream в полиморфный serverStream для инжекции обогащенного контекста
		return handler(srv, &serverStream{
			ServerStream: ss,
			ctx:          transport.WithTraceID(ss.Context(), traceID),
		})
	}
}
