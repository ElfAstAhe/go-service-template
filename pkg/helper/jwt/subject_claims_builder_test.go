package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubjectClaimsBuilder_FluentAPI(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	exp := now.Add(time.Hour)
	nbf := now.Add(-time.Minute)

	t.Run("success fully populated claims building", func(t *testing.T) {
		builder := NewSubjectClaimsBuilder()

		claims := builder.
			WithIssuer("tiny-auth-service").
			WithSubject("test-user").
			WithAudience("gateway", "audit").
			WithID("jti-111").
			WithSubjectID("user-uuid-999").
			WithSubjectType("user").
			WithAdmin(true).
			WithRoles("admin", "manager").
			WithIssuedAt(now).
			WithExpiresAt(exp).
			WithNotBefore(nbf).
			Build()

		require.NotNil(t, claims)
		assert.Equal(t, "tiny-auth-service", claims.Issuer)
		assert.Equal(t, "test-user", claims.Subject)
		assert.Equal(t, jwt.ClaimStrings{"gateway", "audit"}, claims.Audience)
		assert.Equal(t, "jti-111", claims.ID)
		assert.Equal(t, "user-uuid-999", claims.SubjectID)
		assert.Equal(t, "user", claims.SubjectType)
		assert.True(t, claims.Admin)
		assert.Equal(t, []string{"admin", "manager"}, claims.Roles)

		assert.Equal(t, now.Unix(), claims.IssuedAt.Unix())
		assert.Equal(t, exp.Unix(), claims.ExpiresAt.Unix())
		assert.Equal(t, nbf.Unix(), claims.NotBefore.Unix())
	})

	t.Run("zero time values results in nil numeric dates", func(t *testing.T) {
		builder := NewSubjectClaimsBuilder()

		claims := builder.
			WithIssuedAt(time.Time{}).
			WithExpiresAt(time.Time{}).
			WithNotBefore(time.Time{}).
			Build()

		require.NotNil(t, claims)
		assert.Nil(t, claims.IssuedAt)
		assert.Nil(t, claims.ExpiresAt)
		assert.Nil(t, claims.NotBefore)
	})

	t.Run("builder atomically resets state after Build call", func(t *testing.T) {
		builder := NewSubjectClaimsBuilder()

		// Сборка первого экземпляра
		claims1 := builder.
			WithIssuer("issuer-1").
			WithSubjectID("id-1").
			Build()

		// Сборка второго экземпляра тем же билдером
		claims2 := builder.
			WithIssuer("issuer-2").
			WithSubjectID("id-2").
			Build()

		// Проверяем полную изоляцию первого объекта
		require.NotNil(t, claims1)
		assert.Equal(t, "issuer-1", claims1.Issuer)
		assert.Equal(t, "id-1", claims1.SubjectID)

		// Проверяем, что второй объект чист от данных первого
		require.NotNil(t, claims2)
		assert.Equal(t, "issuer-2", claims2.Issuer)
		assert.Equal(t, "id-2", claims2.SubjectID)
		assert.NotEqual(t, claims1, claims2)
	})
}
