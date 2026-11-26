package interceptors

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/auth"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkghlpjwt "github.com/ElfAstAhe/go-service-template/pkg/helper/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// ErrCredentialsNotFound сигнализирует конвейеру, что в gRPC-контексте отсутствуют
// учётные данные для данного способа аутентификации.
var ErrCredentialsNotFound = errors.New("credentials not found in grpc context")

// GRPCCredentialsExtractor извлекает сырые учётные данные из входящего gRPC-контекста.
type GRPCCredentialsExtractor func(ctx context.Context) (any, error)

// GRPCAuthProvider связывает извлечение кредов из gRPC-контекста и их проверку через auth.Authenticator.
type GRPCAuthProvider struct {
	Extract       GRPCCredentialsExtractor
	Authenticator auth.Authenticator
}

// GRPCAuthErrorHandler позволяет кастомизировать преобразование ошибки аутентификации в gRPC status.
type GRPCAuthErrorHandler func(ctx context.Context, err error) error

// AuthUnaryInterceptor создаёт gRPC UnaryServerInterceptor для аутентификации запросов.
func AuthUnaryInterceptor(
	providers []GRPCAuthProvider,
	strict bool,
	errHandler GRPCAuthErrorHandler,
) grpc.UnaryServerInterceptor {
	if errHandler == nil {
		errHandler = defaultGRPCAuthErrorHandler
	}

	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		authCtx, err := authenticateGRPCContext(ctx, providers, strict)
		if err != nil {
			return nil, errHandler(ctx, err)
		}

		return handler(authCtx, req)
	}
}

// AuthStreamInterceptor создаёт gRPC StreamServerInterceptor для аутентификации потоковых вызовов.
func AuthStreamInterceptor(
	providers []GRPCAuthProvider,
	strict bool,
	errHandler GRPCAuthErrorHandler,
) grpc.StreamServerInterceptor {
	if errHandler == nil {
		errHandler = defaultGRPCAuthErrorHandler
	}

	return func(
		srv any,
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		authCtx, err := authenticateGRPCContext(ctx, providers, strict)
		if err != nil {
			return errHandler(ctx, err)
		}

		return handler(srv, &serverStream{
			ServerStream: ss,
			ctx:          authCtx,
		})
	}
}

func authenticateGRPCContext(
	ctx context.Context,
	providers []GRPCAuthProvider,
	strict bool,
) (context.Context, error) {
	// Если субъект уже установлен выше по цепочке — просто пропускаем дальше
	if subj, err := auth.FromContext(ctx); err == nil && subj != nil && subj.IsAuthenticated() {
		return ctx, nil
	}

	for _, p := range providers {
		if p.Extract == nil || p.Authenticator == nil {
			continue
		}

		creds, err := p.Extract(ctx)
		if err != nil {
			if errors.Is(err, ErrCredentialsNotFound) {
				continue
			}
			return nil, err
		}

		subject, err := p.Authenticator.Authenticate(ctx, creds)
		if err != nil {
			return nil, err
		}

		return auth.WithSubject(ctx, subject), nil
	}

	if strict {
		return nil, errs.NewUtlAuthError("authentication credentials required", nil)
	}

	return auth.WithSubject(ctx, auth.Guest), nil
}

func defaultGRPCAuthErrorHandler(_ context.Context, err error) error {
	return status.Error(codes.Unauthenticated, err.Error())
}

// ====================================================================
// Готовые gRPC-экстракторы под JWT, Basic и Cert (mTLS)
// ====================================================================

// NewGRPCJWTExtractor адаптирует pkghlpjwt.GRPCExtractor под конвейер gRPC-аутентификации.
func NewGRPCJWTExtractor(extractor *pkghlpjwt.GRPCExtractor, metadataKey string) GRPCCredentialsExtractor {
	return func(ctx context.Context) (any, error) {
		token, err := extractor.TokenFromContext(ctx, metadataKey)
		if err != nil {
			return nil, ErrCredentialsNotFound
		}

		return &auth.JWTCredentials{
			Token: token,
		}, nil
	}
}

// NewGRPCBasicExtractor извлекает логин и пароль из gRPC metadata ("authorization: Basic <base64>").
func NewGRPCBasicExtractor(metadataKey string) GRPCCredentialsExtractor {
	if strings.TrimSpace(metadataKey) == "" {
		metadataKey = auth.DefaultMetadataName
	}

	return func(ctx context.Context) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrCredentialsNotFound
		}

		values := md.Get(metadataKey)
		if len(values) == 0 {
			return nil, ErrCredentialsNotFound
		}

		username, password, ok := parseBasicAuthHeader(values[0])
		if !ok {
			return nil, ErrCredentialsNotFound
		}

		return &auth.BasicCredentials{
			Username: username,
			Password: password,
		}, nil
	}
}

// NewGRPCCertExtractor извлекает проверенный клиентский TLS-сертификат из gRPC peer.
func NewGRPCCertExtractor() GRPCCredentialsExtractor {
	return func(ctx context.Context) (any, error) {
		p, ok := peer.FromContext(ctx)
		if !ok || p.AuthInfo == nil {
			return nil, ErrCredentialsNotFound
		}

		tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
		if !ok || len(tlsInfo.State.PeerCertificates) == 0 {
			return nil, ErrCredentialsNotFound
		}

		return &auth.CertCredentials{
			Certificate: tlsInfo.State.PeerCertificates[0],
		}, nil
	}
}

func parseBasicAuthHeader(authHeader string) (username, password string, ok bool) {
	const prefix = auth.DefaultBasicPrefix
	if len(authHeader) < len(prefix) || !strings.EqualFold(authHeader[:len(prefix)], prefix) {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
	if err != nil {
		return "", "", false
	}

	cs := string(decoded)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}

	return username, password, true
}
