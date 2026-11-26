package auth

import (
	"context"
	"slices"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CertVerifyFunc резолвит доменный Subject по проверенному Common Name (CN) клиентского сертификата.
type CertVerifyFunc func(ctx context.Context, commonName string) (*Subject, error)

// CertAuthenticator проверяет клиентский X.509-сертификат (mTLS) и собирает Subject.
type CertAuthenticator struct {
	allowedCNs []string
	verify     CertVerifyFunc
}

var _ Authenticator = (*CertAuthenticator)(nil)

func NewCertAuthenticator(allowedCNs []string, verify CertVerifyFunc) (*CertAuthenticator, error) {
	if verify == nil {
		return nil, errs.NewInvalidArgumentError("verify", "is required")
	}
	return &CertAuthenticator{
		allowedCNs: slices.Clone(allowedCNs),
		verify:     verify,
	}, nil
}

func (ca *CertAuthenticator) Authenticate(ctx context.Context, credentials any) (*Subject, error) {
	creds, ok := credentials.(*CertCredentials)
	if !ok || creds.Certificate == nil {
		return nil, errs.NewUtlAuthError("invalid credentials type: expected non-nil CertCredentials", nil)
	}

	cn := strings.TrimSpace(creds.Certificate.Subject.CommonName)
	if cn == "" {
		return nil, errs.NewUtlAuthError("client certificate CommonName is empty", nil)
	}

	if len(ca.allowedCNs) > 0 && !slices.Contains(ca.allowedCNs, cn) {
		return nil, errs.NewUtlAuthError("client certificate CommonName is not allowed", nil)
	}

	subject, err := ca.verify(ctx, cn)
	if err != nil {
		return nil, errs.NewUtlAuthError("certificate authentication failed", err)
	}
	if subject == nil {
		return nil, errs.NewUtlAuthError("certificate authentication returned empty subject", nil)
	}

	return subject, nil
}
