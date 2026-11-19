package jwt

import (
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

func TestNewJWTTokenBuilder(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")

	t.Run("success creation with default id builder", func(t *testing.T) {
		factory, err := NewJWTTokenBuilder("test-issuer", "HS256", secret, time.Hour, nil)

		require.NoError(t, err)
		require.NotNil(t, factory)
		assert.Equal(t, "test-issuer", factory.issuer)
		assert.Equal(t, jwt.SigningMethodHS256, factory.signingMethod)
		assert.NotNil(t, factory.idBuilder)
	})

	t.Run("failed creation with invalid signing method", func(t *testing.T) {
		factory, err := NewJWTTokenBuilder("test-issuer", "INVALID_METHOD", secret, time.Hour, nil)

		var jwtErr *errs.UtlJWTError
		require.Error(t, err)
		assert.ErrorAs(t, err, &jwtErr)
		assert.Nil(t, factory)
	})

	t.Run("success creation with custom mocked id builder", func(t *testing.T) {
		m := new(mockTokenIDBuilder)
		m.On("Generate").Return("mocked-jti-123", nil).Once()

		factory, err := NewJWTTokenBuilder("test-issuer", "HS256", secret, time.Hour, m.Generate)

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
	factory, err := NewJWTTokenBuilder("test-issuer", "HS256", secret, time.Hour, nil)
	require.NoError(t, err)

	t.Run("success build and validate claims", func(t *testing.T) {
		builder := NewSubjectClaimsBuilder().
			WithSubject("vanya").
			WithSubjectType("user").
			WithSubjectID("uuid-123")

		claims, err := factory.Claims(builder)

		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "vanya", claims.Subject)
		assert.Equal(t, "user", claims.SubjectType)
		assert.Equal(t, "uuid-123", claims.SubjectID)
	})

	t.Run("nil builder returns InvalidArgumentError", func(t *testing.T) {
		claims, err := factory.Claims(nil)

		var invalidArgErr *errs.InvalidArgumentError
		require.Error(t, err)
		assert.ErrorAs(t, err, &invalidArgErr)
		assert.Nil(t, claims)
	})

	t.Run("invalid claims validation fails inside factory", func(t *testing.T) {
		builder := NewSubjectClaimsBuilder().WithSubject("") // Пустой субъект вызовет ошибку

		claims, err := factory.Claims(builder)

		require.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestFactory_Token(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")
	factory, err := NewJWTTokenBuilder("test-issuer", "HS256", secret, time.Hour, nil)
	require.NoError(t, err)

	t.Run("success create jwt.Token object", func(t *testing.T) {
		claims := &SubjectClaims{
			SubjectID: "id",
		}

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
	factory, err := NewJWTTokenBuilder("test-issuer", "HS256", secret, time.Hour, nil)
	require.NoError(t, err)

	t.Run("success full signing signature cycle", func(t *testing.T) {
		claims := &SubjectClaims{
			SubjectID: "user-uuid",
		}
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

func TestFactory_DefaultIDBuilder(t *testing.T) {
	secret := []byte("super-secret-key-1234567890123456")
	factory, err := NewJWTTokenBuilder("test-issuer", "HS256", secret, time.Hour, nil)
	require.NoError(t, err)

	t.Run("generate unique uuid with default template prefix", func(t *testing.T) {
		jti1, err1 := factory.defaultIDBuilder()
		jti2, err2 := factory.defaultIDBuilder()

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEmpty(t, jti1)
		assert.NotEmpty(t, jti2)
		assert.NotEqual(t, jti1, jti2)
		assert.Contains(t, jti1, "undef-")
	})
}
