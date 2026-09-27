package helper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
)

// JWTHTTPHelper реализует специализированный транспортный адаптер (HTTP Bridge) для JWT-менеджера.
//
// Инкапсулирует рутину низкоуровневого извлечения сессионных токенов из HTTP-заголовков (Headers)
// и файлов Cookie, позволяя HTTP-middleware прозрачно осуществлять аутентификацию входящего трафика.
type JWTHTTPHelper struct {
	jwtHelper *JWTHelper // Ссылка на базовый криптографический JWT-менеджер фреймворка
}

// NewJWTHTTPHelper — фабричный конструктор HTTP-адаптера авторизации.
func NewJWTHTTPHelper(jwtHelper *JWTHelper) *JWTHTTPHelper {
	return &JWTHTTPHelper{
		jwtHelper: jwtHelper,
	}
}

// ExtractTokenStringFromCookie вытаскивает сырую строку токена из переданного объекта *http.Cookie.
// Производит предварительную валидацию куки на соответствие стандартам спецификации RFC 6265.
func (jhh *JWTHTTPHelper) ExtractTokenStringFromCookie(cookieName string, cookie *http.Cookie) (string, error) {
	if strings.TrimSpace(cookieName) == "" {
		return "", errs.NewInvalidArgumentError("cookieName", "empty cookie name")
	}
	if cookie == nil {
		return "", errs.NewInvalidArgumentError("cookie", "cookie is nil")
	}
	// Верифицируем соответствие параметров куки (имя, печатные символы в значении) стандартам HTTP
	if err := cookie.Valid(); err != nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] is invalid", cookieName), err)
	}

	// Если значение куки передано в чистом виде без сетевого префикса — возвращаем как есть
	if !strings.HasPrefix(cookie.Value, TokenPrefix) {
		return cookie.Value, nil
	}

	return strings.TrimPrefix(cookie.Value, TokenPrefix), nil
}

// ExtractTokenStringFromRequestCookie осуществляет поиск куки по её текстовому имени внутри HTTP-запроса
// и возвращает извлеченную и очищенную от префиксов строку JWT-токена.
func (jhh *JWTHTTPHelper) ExtractTokenStringFromRequestCookie(cookieName string, req *http.Request) (string, error) {
	if strings.TrimSpace(cookieName) == "" {
		return "", errs.NewInvalidArgumentError("cookieName", "empty cookie name")
	}
	if req == nil {
		return "", errs.NewInvalidArgumentError("request", "nil HTTP Request")
	}

	// Извлекаем куку из заголовков запроса. Метод вернет ошибку http.ErrNoCookie, если её нет
	cookie, err := req.Cookie(cookieName)
	if err != nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] extraction", cookieName), err)
	}
	if cookie == nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie not found [%s]", cookieName), nil)
	}

	res, err := jhh.ExtractTokenStringFromCookie(cookieName, cookie)
	if err != nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] value extract", cookieName), err)
	}

	return res, nil
}

// ExtractTokenFromRequestCookie выполняет извлечение строки токена из указанной Cookie HTTP-запроса
// и передает её в базовый JWT-хелпер для полной криптографической верификации цифровой подписи.
func (jhh *JWTHTTPHelper) ExtractTokenFromRequestCookie(cookie *http.Cookie, req *http.Request) (*jwt.Token, error) {
	tokenString, err := jhh.ExtractTokenStringFromRequestCookie(cookie.Name, req)
	if err != nil {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] value extract", cookie.Name), err)
	}

	return jhh.jwtHelper.ExtractTokenFromString(tokenString)
}

// ExtractTokenStringFromHeader вытаскивает строку токена из указанного HTTP-заголовка (карта http.Header).
// Метод строго требует наличия каноничного префикса «Bearer », возвращая пустую строку при его отсутствии.
func (jhh *JWTHTTPHelper) ExtractTokenStringFromHeader(headerName string, headers http.Header) (string, error) {
	if strings.TrimSpace(headerName) == "" {
		return "", errs.NewInvalidArgumentError("headerName", "empty cookie name")
	}
	if headers == nil {
		return "", errs.NewInvalidArgumentError("headers", "nil HTTP Request")
	}

	// Защитный барьер: метод Get возвращает пустую строку, если заголовок отсутствует
	if !strings.HasPrefix(headers.Get(headerName), TokenPrefix) {
		return "", nil
	}

	return strings.TrimPrefix(headers.Get(headerName), TokenPrefix), nil
}

// ExtractTokenStringFromRequestHeader осуществляет поиск и извлечение строки JWT-токена
// напрямую из структуры заголовков входящего HTTP-запроса (*http.Request).
func (jhh *JWTHTTPHelper) ExtractTokenStringFromRequestHeader(headerName string, request *http.Request) (string, error) {
	if strings.TrimSpace(headerName) == "" {
		return "", errs.NewInvalidArgumentError("headerName", "empty cookie name")
	}
	if request == nil {
		return "", errs.NewInvalidArgumentError("request", "nil HTTP Request")
	}
	res, err := jhh.ExtractTokenStringFromHeader(headerName, request.Header)
	if err != nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("header [%s] value extract", headerName), err)
	}

	return res, nil
}

// ExtractTokenFromRequestHeader выполняет извлечение строки токена из целевого HTTP-заголовка запроса
// и передает её в базовый JWT-хелпер для математической проверки подписи и сроков валидности.
func (jhh *JWTHTTPHelper) ExtractTokenFromRequestHeader(headerName string, request *http.Request) (*jwt.Token, error) {
	tokenString, err := jhh.ExtractTokenStringFromHeader(headerName, request.Header)
	if err != nil {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("header [%s] value extract", headerName), err)
	}

	return jhh.jwtHelper.ExtractTokenFromString(tokenString)
}
