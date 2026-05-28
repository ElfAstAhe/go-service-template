package test

import (
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/helper"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTHelper(t *testing.T) {
	secret := "test-secret-key"
	issuer := "test-issuer"

	// Константный ID для тестов
	const fixedTokenID = "fixed-id-123"
	mockIDBuilder := func() string { return fixedTokenID }

	hlp := helper.NewJWTHelper(issuer, jwt.SigningMethodHS256, secret, time.Minute, mockIDBuilder)

	t.Run("Build and Extract: Успешный цикл", func(t *testing.T) {
		// 1. Создаем токен
		tokenStr, err := hlp.BuildTokenString("user-1", "subject", "user", true, "admin")
		require.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		// 2. Парсим обратно
		token, err := hlp.ExtractTokenFromString(tokenStr)
		require.NoError(t, err)
		assert.True(t, token.Valid)

		// 3. Проверяем Claims
		claims, err := hlp.ExtractClaims(token)
		require.NoError(t, err)
		assert.Equal(t, "user-1", claims.SubjectID)
		assert.Equal(t, fixedTokenID, claims.ID) // Проверка нашего IDBuilder
		assert.Equal(t, issuer, claims.Issuer)
		assert.Contains(t, claims.Roles, "admin")
	})

	t.Run("Security: Неверный секретный ключ", func(t *testing.T) {
		tokenStr, _ := hlp.BuildTokenString("1", "s", "t", false)

		evilHelper := helper.NewDefaultJWTHelper("wrong-key")
		_, err := evilHelper.ExtractTokenFromString(tokenStr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature is invalid")
	})

	t.Run("Security: Подмена метода подписи", func(t *testing.T) {
		// Создаем токен методом RS256 (а хелпер ждет HS256)
		claims, _ := hlp.BuildClaims("1", "s", "t", false)
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) // Другой метод HMAC
		tokenStr, _ := token.SignedString([]byte(secret))

		_, err := hlp.ExtractTokenFromString(tokenStr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signing method")
	})

	t.Run("Expiration: Протухший токен", func(t *testing.T) {
		// Хелпер с отрицательным временем жизни
		expiredHelper := helper.NewJWTHelper(issuer, jwt.SigningMethodHS256, secret, -time.Hour, mockIDBuilder)
		tokenStr, _ := expiredHelper.BuildTokenString("1", "s", "t", false)

		_, err := hlp.ExtractTokenFromString(tokenStr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})
}
