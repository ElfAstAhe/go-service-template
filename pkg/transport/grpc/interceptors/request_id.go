package interceptors

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Константы gRPC метаданных для сквозного отслеживания цепочек распределенных вызовов.
const (
	MDXRequestID     string = "x-request-id"     // Стандартный сквозной идентификатор атомарного запроса
	MDXCorrelationID string = "x-correlation-id" // Сквозной идентификатор комплексной бизнес-цепочки (Correlation)
	MDRequestID      string = "x-request-id"     // Дублирующий маркер для обратной совместимости систем
)

// RequestIDExtractorUSInterceptor возвращает готовый унарный интерцептор (Unary Server Interceptor),
// отвечающий за извлечение входящего RequestID или его генерацию с последующей инжекцией в контекст горутины.
func RequestIDExtractorUSInterceptor(headers ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var requestID string
		// Последовательно сканируем переданный приоритетный срез метаданных
		for _, header := range headers {
			vals := metadata.ValueFromIncomingContext(ctx, header)
			if len(vals) > 0 {
				requestID = vals[0]
				break // Извлекаем первый валидный непустой маркер и прерываем цикл (First-Match)
			}
		}

		// Фолбек: если маркер не передан извне — генерируем детерминированный ID на базе префикса и атомарного счетчика
		if requestID == "" {
			requestID = fmt.Sprintf("%s-%07d", transport.GetPrefix(), transport.NextReqID())
		}

		// Передаем управление дальше, упаковав RequestID в context.Context средствами пакета transport
		return handler(transport.WithRequestID(ctx, requestID), req)
	}
}

// RequestIDExtractorSSInterceptor возвращает готовый стриминговый интерцептор (Stream Server Interceptor),
// гарантирующий непрерывное сквозное трассирование и ведение логов для gRPC-потоков (Streams).
func RequestIDExtractorSSInterceptor(headers ...string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var requestID string
		// Сканируем метаданные входящего потока данных
		for _, header := range headers {
			vals := metadata.ValueFromIncomingContext(ss.Context(), header)
			if len(vals) > 0 {
				requestID = vals[0]
				break
			}
		}
		if requestID == "" {
			requestID = fmt.Sprintf("%s-%07d", transport.GetPrefix(), transport.NextReqID())
		}

		// Оборачиваем исходный ServerStream в полиморфный serverStream для инжекции обогащенного контекста
		return handler(srv, &serverStream{
			ServerStream: ss,
			ctx:          transport.WithRequestID(ss.Context(), requestID),
		})
	}
}
