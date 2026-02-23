package helper

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"
)

type JWTGRPCHelper struct {
	jwtHelper *JWTHelper
}

func NewJWTGRPCHelper(jwtHelper *JWTHelper) *JWTGRPCHelper {
	return &JWTGRPCHelper{
		jwtHelper: jwtHelper,
	}
}

func (jgh *JWTGRPCHelper) ExtractTokenStringFromMetadata(metadataName string, md metadata.MD) (string, error) {
	if strings.TrimSpace(metadataName) == "" {
		return "", errs.NewInvalidArgumentError("metadataName", "empty metadata name")
	}
	if md == nil {
		return "", errs.NewInvalidArgumentError("md", "nil metadata")
	}

	values := md.Get(metadataName)
	if len(values) == 0 {
		return "", nil
	}

	if !strings.HasPrefix(values[0], TokenPrefix) {
		return values[0], nil
	}

	return strings.TrimPrefix(values[0], TokenPrefix), nil
}

func (jgh *JWTGRPCHelper) ExtractTokenStringFromContext(metadataName string, ctx context.Context) (string, error) {
	if strings.TrimSpace(metadataName) == "" {
		return "", errs.NewInvalidArgumentError("metadataName", "empty metadata name")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errs.NewInvalidArgumentError("metadata", "no metadata")
	}

	res, err := jgh.ExtractTokenStringFromMetadata(metadataName, md)
	if err != nil {
		return "", errs.NewUtlJWTError("extract token string from metadata", err)
	}

	return res, nil
}

func (jgh *JWTGRPCHelper) ExtractTokenFromContext(metadataName string, ctx context.Context) (*jwt.Token, error) {
	tokenString, err := jgh.ExtractTokenStringFromContext(metadataName, ctx)
	if err != nil {
		return nil, errs.NewUtlJWTError("extract token string from context", err)
	}

	return jgh.jwtHelper.ExtractTokenFromString(tokenString)
}
