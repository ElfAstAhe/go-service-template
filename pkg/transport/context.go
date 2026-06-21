package transport

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// requestIDKey ключ контекста с RequestID
type requestIDKey struct{}

// traceIDKey ключ контекста с TraceID
type traceIDKey struct{}

type realIPKey struct{}

var reqIDCtxKey = requestIDKey{}
var trcIDCtxKey = traceIDKey{}
var realIPCtxKey = realIPKey{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if utils.IsNil(ctx) {
		return ctx
	}

	return context.WithValue(ctx, reqIDCtxKey, requestID)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if utils.IsNil(ctx) {
		return ctx
	}

	return context.WithValue(ctx, trcIDCtxKey, traceID)
}

func WithRealIP(ctx context.Context, realIP string) context.Context {
	if utils.IsNil(ctx) {
		return ctx
	}

	return context.WithValue(ctx, realIPCtxKey, realIP)
}

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
