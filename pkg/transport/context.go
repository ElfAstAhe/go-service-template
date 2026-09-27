package transport

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// Использование уникальных неэкспортируемых пустых структур в качестве ключей
// полностью исключает риск коллизии данных в контексте (Context Collision)
// между независимыми пакетами фреймворка и сторонними библиотеками.
type requestIDKey struct{}
type traceIDKey struct{}
type realIPKey struct{}

// Набор приватных синглтонов-ключей для внутренней работы с context.Context.
var (
	reqIDCtxKey  = requestIDKey{}
	trcIDCtxKey  = traceIDKey{}
	realIPCtxKey = realIPKey{}
)

// WithRequestID обогащает контекст уникальным идентификатором запроса (X-Request-ID).
// Используется в HTTP/gRPC middleware для сквозного отслеживания цепочки логов (Log Correlation).
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if utils.IsNil(ctx) {
		return ctx
	}

	return context.WithValue(ctx, reqIDCtxKey, requestID)
}

// WithTraceID обогащает контекст идентификатором распределенной трассировки OpenTelemetry (Trace ID).
// Позволяет связывать логи и спаны распределенных систем в единый граф вызовов (Jaeger/Zipkin).
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if utils.IsNil(ctx) {
		return ctx
	}

	return context.WithValue(ctx, trcIDCtxKey, traceID)
}

// WithRealIP сохраняет в контексте реальный IP-адрес клиента (X-Real-IP / X-Forwarded-For).
// Используется для систем безопасности, аудит-логов (tiny-audit) и rate-limiting контроля.
func WithRealIP(ctx context.Context, realIP string) context.Context {
	if utils.IsNil(ctx) {
		return ctx
	}

	return context.WithValue(ctx, realIPCtxKey, realIP)
}

// RequestID извлекает идентификатор запроса из контекста.
// Возвращает пустую строку, если ключ не инициализирован или поврежден.
func RequestID(ctx context.Context) string {
	if utils.IsNil(ctx) {
		return ""
	}

	res, ok := ctx.Value(reqIDCtxKey).(string)
	if !ok {
		return ""
	}

	return res
}

// TraceID извлекает идентификатор распределенной трассировки из контекста.
// Безопасно приводит интерфейсный тип к строке с валидацией флага рантайма.
func TraceID(ctx context.Context) string {
	if utils.IsNil(ctx) {
		return ""
	}

	res, ok := ctx.Value(trcIDCtxKey).(string)
	if !ok {
		return ""
	}

	return res
}

// RealIP извлекает сетевой IP-адрес первоначального клиента из контекста.
func RealIP(ctx context.Context) string {
	if utils.IsNil(ctx) {
		return ""
	}

	res, ok := ctx.Value(realIPCtxKey).(string)
	if !ok {
		return ""
	}

	return res
}
