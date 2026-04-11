package interceptors

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	MDXCloudTraceContext string = "x-cloud-trace-context"
	MDTraceParent        string = "traceparent"
	MDXTraceID           string = "x-trace-id"
	MDTraceID            string = "trace-id"
)

func TraceIDExtractorUSInterceptor(headers []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var traceID string
		for _, header := range headers {
			vals := metadata.ValueFromIncomingContext(ctx, header)
			if len(vals) > 0 {
				traceID = vals[0]
				break
			}
		}

		return handler(transport.WithTraceID(ctx, traceID), req)
	}
}

func TraceIDExtractorSSInterceptor(headers []string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var traceID string
		for _, header := range headers {
			vals := metadata.ValueFromIncomingContext(ss.Context(), header)
			if len(vals) > 0 {
				traceID = vals[0]
				break
			}
		}

		return handler(srv, &serverStream{
			ServerStream: ss,
			ctx:          transport.WithTraceID(ss.Context(), traceID),
		})
	}
}
