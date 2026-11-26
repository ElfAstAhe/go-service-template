package auth

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkghlpjwt "github.com/ElfAstAhe/go-service-template/pkg/helper/jwt"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// JWTAuthenticator выполняет криптографическую верификацию сырого JWT-токена
// и преобразует его утверждения (claims) в доменный объект безопасности Subject.
type JWTAuthenticator struct {
	extractor pkghlpjwt.TokenExtractor
}

var _ Authenticator = (*JWTAuthenticator)(nil)

// NewJWTAuthenticator — фабричный конструктор JWT-аутентификатора.
func NewJWTAuthenticator(extractor pkghlpjwt.TokenExtractor) (*JWTAuthenticator, error) {
	if utils.IsNil(extractor) {
		return nil, errs.NewInvalidArgumentError("extractor", "is required")
	}

	return &JWTAuthenticator{
		extractor: extractor,
	}, nil
}

// Authenticate проверяет строковый JWT-токен и возвращает верифицированный Subject.
func (ja *JWTAuthenticator) Authenticate(_ context.Context, credentials any) (*Subject, error) {
	if utils.IsNil(credentials) {
		return nil, errs.NewInvalidArgumentError("credentials", "is required")
	}
	creds, ok := credentials.(*JWTCredentials)
	if !ok {
		return nil, errs.NewUtlAuthError("invalid credentials type: expected JWTCredentials", nil)
	}
	if utils.IsNil(creds.Token) {
		return nil, errs.NewUtlAuthError("empty JWT token", nil)
	}

	claims, err := ja.extractor.Claims(creds.Token)
	if err != nil {
		return nil, errs.NewUtlAuthError("jwt claims extraction failed", err)
	}

	subject, err := MapClaimsToSubject(claims)
	if err != nil {
		return nil, errs.NewUtlAuthError("failed to map jwt claims to subject", err)
	}

	return subject, nil
}
