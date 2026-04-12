package transport

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// requestIDKey ключ контекста с RequestID
type requestIDKey struct{}

// traceIDKey ключ контекста с TraceID
type traceIDKey struct{}

var reqID = requestIDKey{}
var trcID = traceIDKey{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if utils.IsNil(ctx) {
		return ctx
	}

	return context.WithValue(ctx, reqID, requestID)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if utils.IsNil(ctx) {
		return ctx
	}

	return context.WithValue(ctx, trcID, traceID)
}

func RequestID(ctx context.Context) string {
	if utils.IsNil(ctx) {
		return ""
	}

	res, ok := ctx.Value(reqID).(string)
	if !ok {
		return ""
	}

	return res
}

func TraceID(ctx context.Context) string {
	if utils.IsNil(ctx) {
		return ""
	}

	res, ok := ctx.Value(trcID).(string)
	if !ok {
		return ""
	}

	return res
}
