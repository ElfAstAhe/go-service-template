package test

import (
	"net/http"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTHTTPHelper(t *testing.T) {
	baseHelper := helper.NewDefaultJWTHelper("secret")
	httpHelper := helper.NewJWTHTTPHelper(baseHelper)
	headerName := "Authorization"

	t.Run("Header: успешно извлекает токен с Bearer", func(t *testing.T) {
		headers := http.Header{}
		headers.Set(headerName, "Bearer my-token")

		res, err := httpHelper.ExtractTokenStringFromHeader(headers, headerName)

		assert.NoError(t, err)
		assert.Equal(t, "my-token", res)
	})

	t.Run("Cookie: успешно извлекает из структуры Cookie", func(t *testing.T) {
		cookie := &http.Cookie{
			Name:  "session",
			Value: "my-cookie-token",
		}

		res, err := httpHelper.ExtractTokenStringFromCookie(cookie, "session")

		assert.NoError(t, err)
		assert.Equal(t, "my-cookie-token", res)
	})

	t.Run("Request: полный цикл из заголовка запроса", func(t *testing.T) {
		tokenStr, _ := baseHelper.BuildTokenString("1", "sub", "type", false)
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(headerName, "Bearer "+tokenStr)

		token, err := httpHelper.ExtractTokenFromRequestHeader(req, headerName)

		require.NoError(t, err)
		assert.True(t, token.Valid)
	})
}
