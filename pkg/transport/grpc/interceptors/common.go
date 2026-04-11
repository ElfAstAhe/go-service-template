package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}
