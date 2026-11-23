package jwt

import (
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExtractor(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")

	t.Run("success creation", func(t *testing.T) {
		extractor, err := NewExtractor(secret, "HS256", "tiny-audit-service", "tiny-auth-service", time.Minute)

		require.NoError(t, err)
		require.NotNil(t, extractor)
		assert.Equal(t, secret, extractor.secretKey)
		assert.NotNil(t, extractor.claimsValidator)
		assert.Len(t, extractor.tokenParseOpts, 2)
	})

	t.Run("failed - empty secret key", func(t *testing.T) {
		extractor, err := NewExtractor(nil, "HS256", "tiny-audit-service", "tiny-auth-service", time.Minute)

		var invalidArgErr *errs.InvalidArgumentError
		require.Error(t, err)
		assert.ErrorAs(t, err, &invalidArgErr)
		assert.Nil(t, extractor)
	})

	t.Run("failed - unknown signing method", func(t *testing.T) {
		extractor, err := NewExtractor(secret, "UNKNOWN_ALG", "tiny-audit-service", "tiny-auth-service", time.Minute)

		var jwtErr *errs.UtlJWTError
		require.Error(t, err)
		assert.ErrorAs(t, err, &jwtErr)
		assert.Nil(t, extractor)
	})
}

func TestExtractor_Token(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")
	wrongSecret := []byte("wrong-secret-key-1234567890123456")

	extractor, err := NewExtractor(secret, "HS256", "tiny-audit-service", "tiny-auth-service", time.Minute)
	require.NoError(t, err)

	factory, err := NewFactory("tiny-auth-service", []string{"tiny-audit-service"}, "HS256", secret, time.Hour, nil)
	require.NoError(t, err)

	now := time.Now().UTC()

	// Валидный токен
	validClaims, err := factory.Claims(
		NewSubjectClaimsBuilder().
			WithSubject("ivan").
			WithSubjectType("user").
			WithSubjectID("uuid-1"),
	)
	require.NoError(t, err)
	validTokenObj, err := factory.Token(validClaims)
	require.NoError(t, err)
	validTokenStr, err := factory.TokenString(validTokenObj)
	require.NoError(t, err)

	// Просроченный токен (подпись валидна, но exp в прошлом)
	expiredClaims := NewSubjectClaimsBuilder().
		WithIssuer("tiny-auth-service").
		WithAudience("tiny-audit-service").
		WithSubject("ivan").
		WithSubjectType("user").
		WithSubjectID("uuid-1").
		WithIssuedAt(now.Add(-2 * time.Hour)).
		WithNotBefore(now.Add(-2 * time.Hour)).
		WithExpiresAt(now.Add(-time.Hour)).
		Build()
	expiredTokenObj, err := factory.Token(expiredClaims)
	require.NoError(t, err)
	expiredTokenStr, err := factory.TokenString(expiredTokenObj)
	require.NoError(t, err)

	// Токен, подписанный другим алгоритмом (HS512 вместо HS256)
	wrongAlgTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS512, validClaims)
	wrongAlgTokenStr, err := wrongAlgTokenObj.SignedString(secret)
	require.NoError(t, err)

	// Токен, подписанный чужим ключом
	wrongKeyTokenStr, err := validTokenObj.SignedString(wrongSecret)
	require.NoError(t, err)

	tests := []struct {
		name           string
		tokenStr       string
		expectErr      bool
		expectedErrTyp any
	}{
		{
			name:      "success parse valid token",
			tokenStr:  validTokenStr,
			expectErr: false,
		},
		{
			name:      "success parse expired token (claims validation deferred to Claims method)",
			tokenStr:  expiredTokenStr,
			expectErr: false,
		},
		{
			name:           "failed - empty token string",
			tokenStr:       "   ",
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:           "failed - malformed token string",
			tokenStr:       "not.a.jwt",
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:           "failed - algorithm confusion (HS512 instead of HS256)",
			tokenStr:       wrongAlgTokenStr,
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:           "failed - invalid signature key",
			tokenStr:       wrongKeyTokenStr,
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, parseErr := extractor.Token(tt.tokenStr)

			if tt.expectErr {
				require.Error(t, parseErr)
				assert.ErrorAs(t, parseErr, tt.expectedErrTyp)
				assert.Nil(t, token)
				return
			}

			require.NoError(t, parseErr)
			require.NotNil(t, token)
			assert.True(t, token.Valid)
		})
	}
}

func TestExtractor_Claims(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")
	extractor, err := NewExtractor(secret, "HS256", "tiny-audit-service", "tiny-auth-service", time.Minute)
	require.NoError(t, err)

	now := time.Now().UTC()

	makeToken := func(claims jwt.Claims) *jwt.Token {
		return &jwt.Token{
			Method: jwt.SigningMethodHS256,
			Claims: claims,
			Valid:  true,
		}
	}

	tests := []struct {
		name           string
		token          *jwt.Token
		expectErr      bool
		expectedErrTyp any
	}{
		{
			name: "success extract and validate claims",
			token: makeToken(
				NewSubjectClaimsBuilder().
					WithIssuer("tiny-auth-service").
					WithAudience("tiny-audit-service").
					WithSubject("ivan").
					WithSubjectType("user").
					WithSubjectID("uuid-1").
					WithRoles("admin").
					WithIssuedAt(now).
					WithNotBefore(now).
					WithExpiresAt(now.Add(time.Hour)).
					Build(),
			),
			expectErr: false,
		},
		{
			name: "success extract guest claims without subject_id",
			token: makeToken(
				NewSubjectClaimsBuilder().
					WithIssuer("tiny-auth-service").
					WithAudience("tiny-audit-service").
					WithSubject("anonymous").
					WithSubjectType("guest").
					WithIssuedAt(now).
					WithNotBefore(now).
					WithExpiresAt(now.Add(time.Hour)).
					Build(),
			),
			expectErr: false,
		},
		{
			name:           "failed - nil token",
			token:          nil,
			expectErr:      true,
			expectedErrTyp: new(*errs.InvalidArgumentError),
		},
		{
			name:           "failed - nil token claims",
			token:          makeToken(nil),
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name:           "failed - unsupported claims type (MapClaims)",
			token:          makeToken(jwt.MapClaims{"sub": "ivan"}),
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name: "failed - expired token",
			token: makeToken(
				NewSubjectClaimsBuilder().
					WithIssuer("tiny-auth-service").
					WithAudience("tiny-audit-service").
					WithSubject("ivan").
					WithSubjectType("user").
					WithSubjectID("uuid-1").
					WithIssuedAt(now.Add(-2 * time.Hour)).
					WithNotBefore(now.Add(-2 * time.Hour)).
					WithExpiresAt(now.Add(-5 * time.Minute)).
					Build(),
			),
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name: "failed - not before in future",
			token: makeToken(
				NewSubjectClaimsBuilder().
					WithIssuer("tiny-auth-service").
					WithAudience("tiny-audit-service").
					WithSubject("ivan").
					WithSubjectType("user").
					WithSubjectID("uuid-1").
					WithIssuedAt(now).
					WithNotBefore(now.Add(5 * time.Minute)).
					WithExpiresAt(now.Add(time.Hour)).
					Build(),
			),
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name: "failed - wrong audience",
			token: makeToken(
				NewSubjectClaimsBuilder().
					WithIssuer("tiny-auth-service").
					WithAudience("billing-service").
					WithSubject("ivan").
					WithSubjectType("user").
					WithSubjectID("uuid-1").
					WithIssuedAt(now).
					WithNotBefore(now).
					WithExpiresAt(now.Add(time.Hour)).
					Build(),
			),
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name: "failed - wrong issuer",
			token: makeToken(
				NewSubjectClaimsBuilder().
					WithIssuer("evil-auth-service").
					WithAudience("tiny-audit-service").
					WithSubject("ivan").
					WithSubjectType("user").
					WithSubjectID("uuid-1").
					WithIssuedAt(now).
					WithNotBefore(now).
					WithExpiresAt(now.Add(time.Hour)).
					Build(),
			),
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
		{
			name: "failed - missing domain subject_type",
			token: makeToken(
				NewSubjectClaimsBuilder().
					WithIssuer("tiny-auth-service").
					WithAudience("tiny-audit-service").
					WithSubject("ivan").
					WithSubjectType("").
					WithIssuedAt(now).
					WithNotBefore(now).
					WithExpiresAt(now.Add(time.Hour)).
					Build(),
			),
			expectErr:      true,
			expectedErrTyp: new(*errs.UtlJWTError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, extractErr := extractor.Claims(tt.token)

			if tt.expectErr {
				require.Error(t, extractErr)
				assert.ErrorAs(t, extractErr, tt.expectedErrTyp)
				assert.Nil(t, claims)
				return
			}

			require.NoError(t, extractErr)
			require.NotNil(t, claims)
		})
	}
}
