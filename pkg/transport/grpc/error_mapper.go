package grpc

import (
	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapToGrpcError(err error) error {
	if err == nil {
		return nil
	}

	// Bad Request
	if transport.IsBadRequest(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Unauthorized
	if transport.IsUnauthorized(err) {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	// Forbidden
	if transport.IsForbidden(err) {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	// Not Found
	if transport.IsNotFound(err) {
		return status.Error(codes.NotFound, err.Error())
	}

	// Conflict
	if transport.IsConflict(err) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	// Gone
	if transport.IsGone(err) {
		return status.Error(codes.NotFound, err.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
