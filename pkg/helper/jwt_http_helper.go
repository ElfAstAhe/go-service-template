package helper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v4"
)

type JWTHTTPHelper struct {
	jwtHelper *JWTHelper
}

func NewJWTHTTPHelper(jwtHelper *JWTHelper) *JWTHTTPHelper {
	return &JWTHTTPHelper{
		jwtHelper: jwtHelper,
	}
}

func (jhh *JWTHTTPHelper) ExtractTokenStringFromCookie(cookieName string, cookie *http.Cookie) (string, error) {
	if strings.TrimSpace(cookieName) == "" {
		return "", errs.NewInvalidArgumentError("cookieName", "empty cookie name")
	}
	if cookie == nil {
		return "", errs.NewInvalidArgumentError("cookie", "cookie is nil")
	}
	if err := cookie.Valid(); err != nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] is invalid", cookieName), err)
	}

	if !strings.HasPrefix(cookie.Value, TokenPrefix) {
		return cookie.Value, nil
	}

	return strings.TrimPrefix(cookie.Value, TokenPrefix), nil
}

func (jhh *JWTHTTPHelper) ExtractTokenStringFromRequestCookie(cookieName string, req *http.Request) (string, error) {
	if strings.TrimSpace(cookieName) == "" {
		return "", errs.NewInvalidArgumentError("cookieName", "empty cookie name")
	}
	if req == nil {
		return "", errs.NewInvalidArgumentError("request", "nil HTTP Request")
	}
	cookie, err := req.Cookie(cookieName)
	if err != nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] extraction", cookieName), err)
	}
	if cookie == nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie not found [%s]", cookieName), err)
	}

	res, err := jhh.ExtractTokenStringFromCookie(cookieName, cookie)
	if err != nil {
		return "", errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] value extract", cookieName), err)
	}

	return res, nil
}

func (jhh *JWTHTTPHelper) ExtractTokenFromRequestCookie(cookie *http.Cookie, req *http.Request) (*jwt.Token, error) {
	tokenString, err := jhh.ExtractTokenStringFromRequestCookie(cookie.Name, req)
	if err != nil {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("cookie [%s] value extract", cookie.Name), err)
	}

	return jhh.jwtHelper.ExtractTokenFromString(tokenString)
}

func (jhh *JWTHTTPHelper) ExtractTokenStringFromHeader(headerName string, headers http.Header) (string, error) {
	if strings.TrimSpace(headerName) == "" {
		return "", errs.NewInvalidArgumentError("headerName", "empty cookie name")
	}
	if headers == nil {
		return "", errs.NewInvalidArgumentError("headers", "nil HTTP Request")
	}

	if !strings.HasPrefix(headers.Get(headerName), TokenPrefix) {
		return "", nil
	}

	return strings.TrimPrefix(headers.Get(headerName), TokenPrefix), nil
}

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

func (jhh *JWTHTTPHelper) ExtractTokenFromRequestHeader(headerName string, request *http.Request) (*jwt.Token, error) {
	tokenString, err := jhh.ExtractTokenStringFromHeader(headerName, request.Header)
	if err != nil {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("header [%s] value extract", headerName), err)
	}

	return jhh.jwtHelper.ExtractTokenFromString(tokenString)
}
