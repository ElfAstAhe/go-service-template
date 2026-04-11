package interceptors

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	MDXRequestID     string = "x-request-id"
	MDXCorrelationID string = "x-correlation-id"
	MDRequestID      string = "x-request-id"
)

func RequestIDExtractorUSInterceptor(headers []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var requestID string
		for _, header := range headers {
			vals := metadata.ValueFromIncomingContext(ctx, header)
			if len(vals) > 0 {
				requestID = vals[0]
				break
			}
		}
		if requestID == "" {
			requestID = fmt.Sprintf("%s-%07d", transport.GetPrefix(), transport.NextReqID())
		}

		return handler(transport.WithRequestID(ctx, requestID), req)
	}
}

func RequestIDExtractorSSInterceptor(headers []string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var requestID string
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

		return handler(srv, &serverStream{
			ServerStream: ss,
			ctx:          transport.WithRequestID(ss.Context(), requestID),
		})
	}
}
