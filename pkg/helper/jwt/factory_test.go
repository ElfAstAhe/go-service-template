package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockTokenIDBuilder симулирует поведение TokenIDBuilderFunc с помощью встроенного движка testify/mock.
type mockTokenIDBuilder struct {
	mock.Mock
}

func (m *mockTokenIDBuilder) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestNewFactory(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")
	audiences := []string{"tiny-audit-service", "billing-service"}

	t.Run("success creation with default id builder", func(t *testing.T) {
		factory, err := NewFactory("tiny-auth-service", audiences, "HS256", secret, time.Hour, nil)

		require.NoError(t, err)
		require.NotNil(t, factory)
		assert.Equal(t, "tiny-auth-service", factory.issuer)
		assert.Equal(t, audiences, factory.audiences)
		assert.Equal(t, jwt.SigningMethodHS256, factory.signingMethod)
		assert.NotNil(t, factory.idBuilder)
	})

	t.Run("failed creation with invalid signing method", func(t *testing.T) {
		factory, err := NewFactory("tiny-auth-service", audiences, "INVALID_METHOD", secret, time.Hour, nil)

		var jwtErr *errs.UtlJWTError
		require.Error(t, err)
		assert.ErrorAs(t, err, &jwtErr)
		assert.Nil(t, factory)
	})

	t.Run("success creation with custom id builder", func(t *testing.T) {
		m := new(mockTokenIDBuilder)
		m.On("Generate").Return("mocked-jti-123", nil).Once()

		factory, err := NewFactory("tiny-auth-service", audiences, "HS256", secret, time.Hour, m.Generate)

		require.NoError(t, err)
		require.NotNil(t, factory)

		jti, err := factory.idBuilder()
		require.NoError(t, err)
		assert.Equal(t, "mocked-jti-123", jti)
		m.AssertExpectations(t)
	})
}

func TestFactory_Claims(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")
	audiences := []string{"tiny-audit-service", "billing-service"}
	ttl := time.Hour

	t.Run("success enrich and validate claims", func(t *testing.T) {
		m := new(mockTokenIDBuilder)
		m.On("Generate").Return("jti-777", nil).Once()

		factory, err := NewFactory("tiny-auth-service", audiences, "HS256", secret, ttl, m.Generate)
		require.NoError(t, err)

		builder := NewSubjectClaimsBuilder().
			WithSubject("ivan").
			WithSubjectType("user").
			WithSubjectID("uuid-123").
			WithRoles("audit:reader")

		before := time.Now().Truncate(time.Second)
		claims, err := factory.Claims(builder)
		after := time.Now().Add(time.Second).Truncate(time.Second)

		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "tiny-auth-service", claims.Issuer)
		assert.Equal(t, jwt.ClaimStrings(audiences), claims.Audience)
		assert.Equal(t, "jti-777", claims.ID)
		assert.Equal(t, "ivan", claims.Subject)
		assert.Equal(t, "user", claims.SubjectType)
		assert.Equal(t, "uuid-123", claims.SubjectID)
		assert.Equal(t, []string{"audit:reader"}, claims.Roles)

		require.NotNil(t, claims.IssuedAt)
		require.NotNil(t, claims.NotBefore)
		require.NotNil(t, claims.ExpiresAt)
		assert.Equal(t, claims.IssuedAt.Unix(), claims.NotBefore.Unix())
		assert.GreaterOrEqual(t, claims.IssuedAt.Unix(), before.Unix())
		assert.LessOrEqual(t, claims.IssuedAt.Unix(), after.Unix())
		assert.Equal(t, claims.IssuedAt.Add(ttl).Unix(), claims.ExpiresAt.Unix())

		m.AssertExpectations(t)
	})

	t.Run("nil builder returns InvalidArgumentError", func(t *testing.T) {
		factory, err := NewFactory("tiny-auth-service", audiences, "HS256", secret, ttl, nil)
		require.NoError(t, err)

		claims, err := factory.Claims(nil)

		var invalidArgErr *errs.InvalidArgumentError
		require.Error(t, err)
		assert.ErrorAs(t, err, &invalidArgErr)
		assert.Nil(t, claims)
	})

	t.Run("idBuilder error returns UtlJWTError", func(t *testing.T) {
		m := new(mockTokenIDBuilder)
		m.On("Generate").Return("", errors.New("entropy failure")).Once()

		factory, err := NewFactory("tiny-auth-service", audiences, "HS256", secret, ttl, m.Generate)
		require.NoError(t, err)

		builder := NewSubjectClaimsBuilder().
			WithSubject("ivan").
			WithSubjectType("user")

		claims, err := factory.Claims(builder)

		var jwtErr *errs.UtlJWTError
		require.Error(t, err)
		assert.ErrorAs(t, err, &jwtErr)
		assert.Nil(t, claims)
		m.AssertExpectations(t)
	})

	t.Run("invalid subject validation fails inside factory", func(t *testing.T) {
		factory, err := NewFactory("tiny-auth-service", audiences, "HS256", secret, ttl, nil)
		require.NoError(t, err)

		builder := NewSubjectClaimsBuilder().WithSubject("")

		claims, err := factory.Claims(builder)

		var jwtErr *errs.UtlJWTError
		require.Error(t, err)
		assert.ErrorAs(t, err, &jwtErr)
		assert.Nil(t, claims)
	})
}

func TestFactory_Token(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")
	factory, err := NewFactory("tiny-auth-service", []string{"audit"}, "HS256", secret, time.Hour, nil)
	require.NoError(t, err)

	t.Run("success create jwt.Token object", func(t *testing.T) {
		claims := &SubjectClaims{SubjectID: "id"}

		token, err := factory.Token(claims)

		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, jwt.SigningMethodHS256, token.Method)
		assert.Equal(t, claims, token.Claims)
	})

	t.Run("nil claims returns InvalidArgumentError", func(t *testing.T) {
		token, err := factory.Token(nil)

		var invalidArgErr *errs.InvalidArgumentError
		require.Error(t, err)
		assert.ErrorAs(t, err, &invalidArgErr)
		assert.Nil(t, token)
	})
}

func TestFactory_TokenString(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")
	factory, err := NewFactory("tiny-auth-service", []string{"audit"}, "HS256", secret, time.Hour, nil)
	require.NoError(t, err)

	t.Run("success full signing cycle", func(t *testing.T) {
		claims := &SubjectClaims{SubjectID: "user-uuid"}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenStr, err := factory.TokenString(token)

		require.NoError(t, err)
		require.NotEmpty(t, tokenStr)
		assert.Contains(t, tokenStr, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	})

	t.Run("nil token reference returns InvalidArgumentError", func(t *testing.T) {
		tokenStr, err := factory.TokenString(nil)

		var invalidArgErr *errs.InvalidArgumentError
		require.Error(t, err)
		assert.ErrorAs(t, err, &invalidArgErr)
		assert.Empty(t, tokenStr)
	})
}
