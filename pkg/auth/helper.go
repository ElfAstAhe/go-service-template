package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/helper"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

// Рантайм-константы имен заголовков, кук и gRPC-метаданных по умолчанию.
const (
	DefaultHeaderName   string = "Authorization"
	DefaultCookieName   string = "Authorization"
	DefaultMetadataName string = "Authorization"
)

// Helper описывает высокоуровневый центральный контракт подсистемы аутентификации (Auth Facade).
type Helper interface {
	SubjectFromToken(token *jwt.Token) (*Subject, error)
	SubjectFromTokenString(tokenString string) (*Subject, error)
	SubjectFromContext(ctx context.Context) (*Subject, error)
	SubjectFromHTTPRequest(request *http.Request) (*Subject, error)
	SubjectFromGRPCMetadata(md metadata.MD) (*Subject, error)
	SubjectFromGRPCContext(gRPCCtx context.Context) (*Subject, error)
	HasSubjectInContext(ctx context.Context) bool
	TokenFromSubject(subject *Subject) (*jwt.Token, error)
	TokenStringFromSubjet(subject *Subject) (string, error)
}

// HelperImpl реализует интерфейс Helper, координируя низкоуровневые транспортные JWT-хелперы фреймворка.
type HelperImpl struct {
	headerName    string
	cookieName    string
	metadataName  string
	jwtHelper     *helper.JWTHelper
	jwtHTTPHelper *helper.JWTHTTPHelper
	jwtGRPCHelper *helper.JWTGRPCHelper
}

// Гарантируем полное соответствие интерфейсу Helper на этапе компиляции
var _ Helper = (*HelperImpl)(nil)

// NewHelper — универсальный конструктор хелпера авторизации с кастомизацией сетевых имен.
func NewHelper(
	headerName string,
	cookieName, metadataName string,
	jwtHelper *helper.JWTHelper,
	jwtHTTPHelper *helper.JWTHTTPHelper,
	jwtGRPCHelper *helper.JWTGRPCHelper,
) *HelperImpl {
	return &HelperImpl{
		headerName:    headerName,
		cookieName:    cookieName,
		metadataName:  metadataName,
		jwtHelper:     jwtHelper,
		jwtHTTPHelper: jwtHTTPHelper,
		jwtGRPCHelper: jwtGRPCHelper,
	}
}

// NewDefaultHelper собирает хелпер авторизации, наполняя его дефолтными JWT-компонентами фреймворка.
func NewDefaultHelper(secretKey string) *HelperImpl {
	jwtHelper := helper.NewDefaultJWTHelper(secretKey)
	jwtHTTPHelper := helper.NewJWTHTTPHelper(jwtHelper)
	jwtGRPCHelper := helper.NewJWTGRPCHelper(jwtHelper)

	return NewDefaultHelperEx(jwtHelper, jwtHTTPHelper, jwtGRPCHelper)
}

// NewDefaultHelperEx собирает хелпер на базе уже существующих инстансов транспортных адаптеров.
func NewDefaultHelperEx(
	jwtHelper *helper.JWTHelper,
	jwtHTTPHelper *helper.JWTHTTPHelper,
	jwtGRPCHelper *helper.JWTGRPCHelper,
) *HelperImpl {
	return NewHelper(DefaultHeaderName, DefaultCookieName, DefaultMetadataName, jwtHelper, jwtHTTPHelper, jwtGRPCHelper)
}

// SubjectFromToken распаковывает утверждения (Claims) валидного JWT-объекта и маппит их в доменную структуру Subject.
func (ah *HelperImpl) SubjectFromToken(token *jwt.Token) (*Subject, error) {
	if token == nil {
		return nil, errs.NewInvalidArgumentError("token", "nil jwt token")
	}

	claims, err := ah.jwtHelper.ExtractClaims(token)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract claims", err)
	}

	return NewSubject(claims.SubjectID, claims.Subject, SubjectType(claims.SubjectType), claims.Roles, nil), nil
}

// TokenFromSubject осуществляет обратную операцию: упаковывает доменный Subject в неподписанный объект jwt.Token.
func (ah *HelperImpl) TokenFromSubject(subject *Subject) (*jwt.Token, error) {
	if subject == nil {
		return nil, errs.NewInvalidArgumentError("subject", "nil user info")
	}

	return ah.jwtHelper.BuildToken(subject.ID, subject.Name, string(subject.Type), false, ah.tokenRolesFromSubject(subject)...)
}

// tokenRolesFromSubject — внутренний хелпер вычленения ключей мапы ролей в плоский строковый срез без лишних аллокаций.
func (ah *HelperImpl) tokenRolesFromSubject(subject *Subject) []string {
	res := make([]string, 0, len(subject.Roles))
	for key := range subject.Roles {
		res = append(res, key)
	}

	return res
}

// TokenStringFromSubjet генерирует токен на базе Subject и сразу подписывает его, возвращая компактную финальную строку.
func (ah *HelperImpl) TokenStringFromSubjet(subject *Subject) (string, error) {
	token, err := ah.TokenFromSubject(subject)
	if err != nil {
		return "", err
	}

	return ah.jwtHelper.BuildTokenStr(token)
}

// SubjectFromTokenString парсит сырую строку JWT, верифицирует подпись и возвращает доменный Subject.
func (ah *HelperImpl) SubjectFromTokenString(tokenString string) (*Subject, error) {
	token, err := ah.jwtHelper.ExtractTokenFromString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract token", err)
	}

	return ah.SubjectFromToken(token)
}

// SubjectFromContext пытается извлечь Subject из context.Context.
// После исправления FromContext метод теперь работает корректно и возвращает ошибку, если сессия пуста.
func (ah *HelperImpl) SubjectFromContext(ctx context.Context) (*Subject, error) {
	res := FromContext(ctx)
	if res != nil {
		return res, nil
	}

	return nil, errs.NewUtlAuthError("user info not found", nil)
}

// HasSubjectInContext проверяет наличие авторизованного субъекта в контексте.
func (ah *HelperImpl) HasSubjectInContext(ctx context.Context) bool {
	userInfo, err := ah.SubjectFromContext(ctx)
	if err != nil {
		return false
	}

	return userInfo != nil
}

// SubjectFromHTTPRequest выполняет двухканальное извлечение JWT (Cookie/Header) из входящего HTTP-запроса REST ручек.
func (ah *HelperImpl) SubjectFromHTTPRequest(request *http.Request) (*Subject, error) {
	cookieTokenString, cookieErr := ah.jwtHTTPHelper.ExtractTokenStringFromRequestCookie(request, ah.cookieName)
	headerTokenString, headerErr := ah.jwtHTTPHelper.ExtractTokenStringFromRequestHeader(request, ah.headerName)
	if cookieErr != nil && headerErr != nil {
		return nil, errs.NewUtlAuthError("extract token string", errors.Join(cookieErr, headerErr))
	}

	var tokenString string
	if cookieErr == nil && cookieTokenString != "" {
		tokenString = cookieTokenString
	} else if headerErr == nil && headerTokenString != "" {
		tokenString = headerTokenString
	}

	userInfo, err := ah.SubjectFromTokenString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract user info", err)
	}

	return userInfo, nil
}

// SubjectFromGRPCMetadata извлекает и валидирует токен напрямую из переданной карты gRPC-метаданных (HTTP/2 заголовков).
func (ah *HelperImpl) SubjectFromGRPCMetadata(md metadata.MD) (*Subject, error) {
	tokenString, err := ah.jwtGRPCHelper.ExtractTokenStringFromMetadata(md, ah.metadataName)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract token string", err)
	}

	userInfo, err := ah.SubjectFromTokenString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract user info", err)
	}

	return userInfo, nil
}

// SubjectFromGRPCContext извлекает и валидирует токен из входящего gRPC контекста context.Context.
func (ah *HelperImpl) SubjectFromGRPCContext(gRPCCtx context.Context) (*Subject, error) {
	tokenString, err := ah.jwtGRPCHelper.ExtractTokenStringFromContext(gRPCCtx, ah.metadataName)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract token string", err)
	}

	userInfo, err := ah.SubjectFromTokenString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract user info", err)
	}

	return userInfo, nil
}
