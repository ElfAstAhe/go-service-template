package auth

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// BasicVerifyFunc определяет функцию проверки пары логин/пароль (локально или через внешний Auth-сервис/кэш).
type BasicVerifyFunc func(ctx context.Context, username, password string) (*Subject, error)

// BasicAuthenticator проверяет учётные данные Basic Auth и возвращает верифицированный Subject.
type BasicAuthenticator struct {
	verify BasicVerifyFunc
}

var _ Authenticator = (*BasicAuthenticator)(nil)

func NewBasicAuthenticator(verify BasicVerifyFunc) (*BasicAuthenticator, error) {
	if verify == nil {
		return nil, errs.NewInvalidArgumentError("verify", "is required")
	}

	return &BasicAuthenticator{verify: verify}, nil
}

func (ba *BasicAuthenticator) Authenticate(ctx context.Context, credentials any) (*Subject, error) {
	creds, ok := credentials.(*BasicCredentials)
	if !ok {
		return nil, errs.NewUtlAuthError("invalid credentials type: expected BasicCredentials", nil)
	}
	if strings.TrimSpace(creds.Username) == "" || creds.Password == "" {
		return nil, errs.NewUtlAuthError("username and password are required", nil)
	}

	subject, err := ba.verify(ctx, creds.Username, creds.Password)
	if err != nil {
		return nil, errs.NewUtlAuthError("basic authentication failed", err)
	}
	if utils.IsNil(subject) {
		return nil, errs.NewUtlAuthError("basic authentication returned empty subject", nil)
	}

	return subject, nil
}
