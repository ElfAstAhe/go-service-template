package jwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// DefaultHTTPTokenPrefix префикс токена по умолчанию
	DefaultHTTPTokenPrefix = "Bearer "
)

// HTTPExtractor инкапсулирует логику извлечения и первичного парсинга JWT-токенов
// из заголовков и cookie HTTP-запросов.
type HTTPExtractor struct {
	tokenExtractor TokenExtractor
	tokenPrefix    string
}

// NewHTTPExtractor создает новый экземпляр HTTPExtractor с заданным экстрактором токенов и префиксом (например, "Bearer").
func NewHTTPExtractor(
	tokenExtractor TokenExtractor,
	tokenPrefix string,
) *HTTPExtractor {
	return &HTTPExtractor{
		tokenExtractor: tokenExtractor,
		tokenPrefix:    tokenPrefix,
	}
}

// TokenFromRequestCookie извлекает JWT-токен из именованной cookie HTTP-запроса и проверяет его подпись.
func (he *HTTPExtractor) TokenFromRequestCookie(req *http.Request, name string) (*jwt.Token, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errs.NewInvalidArgumentError("name", "is required")
	}
	if utils.IsNil(req) {
		return nil, errs.NewInvalidArgumentError("req", "is required")
	}

	cookie, err := req.Cookie(name)
	if err != nil {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] extraction", name), err)
	}
	if utils.IsNil(cookie) {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("cookie not found [%s]", name), nil)
	}
	tokenStr, err := he.tokenStrFromCookie(cookie)
	if err != nil {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] value extract failed", name), err)
	}
	res, err := he.tokenExtractor.Token(tokenStr)
	if err != nil {
		return nil, errs.NewUtlJWTError("token extraction failed", err)
	}

	return res, nil
}

// TokenFromRequestHeader извлекает JWT-токен из указанного заголовка HTTP-запроса и проверяет его подпись.
func (he *HTTPExtractor) TokenFromRequestHeader(req *http.Request, name string) (*jwt.Token, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errs.NewInvalidArgumentError("name", "is required")
	}
	if utils.IsNil(req) {
		return nil, errs.NewInvalidArgumentError("req", "is required")
	}
	tokenStr, err := he.tokenStrFromHeader(req.Header, name)
	if err != nil {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("header [%s] value extract failed", name), err)
	}
	res, err := he.tokenExtractor.Token(tokenStr)
	if err != nil {
		return nil, errs.NewUtlJWTError("token extraction failed", err)
	}

	return res, nil
}

func (he *HTTPExtractor) tokenStrFromCookie(cookie *http.Cookie) (string, error) {
	if utils.IsNil(cookie) {
		return "", errs.NewInvalidArgumentError("cookies", "is required")
	}
	// Верифицируем соответствие параметров куки (имя, печатные символы в значении) стандартам HTTP
	if err := cookie.Valid(); err != nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] is invalid", cookie.Name), err)
	}
	val := cookie.Value
	if strings.TrimSpace(val) == "" {
		return "", errs.NewUtlJWTError("no token data", nil)
	}
	// Если значение куки передано в чистом виде без сетевого префикса — возвращаем как есть
	if !strings.HasPrefix(val, he.tokenPrefix) {
		return val, nil
	}

	return strings.TrimSpace(strings.TrimPrefix(val, he.tokenPrefix)), nil
}

func (he *HTTPExtractor) tokenStrFromHeader(headers http.Header, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errs.NewInvalidArgumentError("name", "is required")
	}
	if utils.IsNil(headers) {
		return "", errs.NewInvalidArgumentError("headers", "is required")
	}

	val := headers.Get(name)
	if strings.TrimSpace(val) == "" {
		return "", errs.NewUtlJWTError("no token data", nil)
	}

	// Защитный барьер: метод Get возвращает пустую строку, если заголовок отсутствует
	if !strings.HasPrefix(val, he.tokenPrefix) {
		return "", nil
	}

	return strings.TrimSpace(strings.TrimPrefix(val, he.tokenPrefix)), nil
}
