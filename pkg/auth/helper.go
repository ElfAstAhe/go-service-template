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

const (
	DefaultHeaderName   string = "Authorization"
	DefaultCookieName   string = "Authorization"
	DefaultMetadataName string = "Authorization"
)

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

type HelperImpl struct {
	headerName    string
	cookieName    string
	metadataName  string
	jwtHelper     *helper.JWTHelper
	jwtHTTPHelper *helper.JWTHTTPHelper
	jwtGRPCHelper *helper.JWTGRPCHelper
}

var _ Helper = (*HelperImpl)(nil)

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

func NewDefaultHelper(secretKey string) *HelperImpl {
	jwtHelper := helper.NewDefaultJWTHelper(secretKey)
	jwtHTTPHelper := helper.NewJWTHTTPHelper(jwtHelper)
	jwtGRPCHelper := helper.NewJWTGRPCHelper(jwtHelper)

	return NewDefaultHelperEx(jwtHelper, jwtHTTPHelper, jwtGRPCHelper)
}

func NewDefaultHelperEx(
	jwtHelper *helper.JWTHelper,
	jwtHTTPHelper *helper.JWTHTTPHelper,
	jwtGRPCHelper *helper.JWTGRPCHelper,
) *HelperImpl {
	return NewHelper(DefaultHeaderName, DefaultCookieName, DefaultMetadataName, jwtHelper, jwtHTTPHelper, jwtGRPCHelper)
}

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

func (ah *HelperImpl) TokenFromSubject(subject *Subject) (*jwt.Token, error) {
	if subject == nil {
		return nil, errs.NewInvalidArgumentError("subject", "nil user info")
	}

	return ah.jwtHelper.BuildToken(subject.ID, subject.Name, string(subject.Type), false, ah.tokenRolesFromSubject(subject)...)
}

func (ah *HelperImpl) tokenRolesFromSubject(subject *Subject) []string {
	res := make([]string, 0, len(subject.Roles))
	for key := range subject.Roles {
		res = append(res, key)
	}

	return res
}

func (ah *HelperImpl) TokenStringFromSubjet(subject *Subject) (string, error) {
	token, err := ah.TokenFromSubject(subject)
	if err != nil {
		return "", err
	}

	return ah.jwtHelper.BuildTokenStr(token)
}

func (ah *HelperImpl) SubjectFromTokenString(tokenString string) (*Subject, error) {
	token, err := ah.jwtHelper.ExtractTokenFromString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract token", err)
	}

	return ah.SubjectFromToken(token)
}

func (ah *HelperImpl) SubjectFromContext(ctx context.Context) (*Subject, error) {
	res := FromContext(ctx)
	if res != nil {
		return res, nil
	}

	return nil, errs.NewUtlAuthError("user info not found", nil)
}

func (ah *HelperImpl) HasSubjectInContext(ctx context.Context) bool {
	userInfo, err := ah.SubjectFromContext(ctx)
	if err != nil {
		return false
	}

	return userInfo != nil
}

func (ah *HelperImpl) SubjectFromHTTPRequest(request *http.Request) (*Subject, error) {
	cookieTokenString, cookieErr := ah.jwtHTTPHelper.ExtractTokenStringFromRequestCookie(ah.cookieName, request)
	headerTokenString, headerErr := ah.jwtHTTPHelper.ExtractTokenStringFromRequestHeader(ah.headerName, request)
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

func (ah *HelperImpl) SubjectFromGRPCMetadata(md metadata.MD) (*Subject, error) {
	tokenString, err := ah.jwtGRPCHelper.ExtractTokenStringFromMetadata(ah.metadataName, md)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract token string", err)
	}

	userInfo, err := ah.SubjectFromTokenString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract user info", err)
	}

	return userInfo, nil
}

func (ah *HelperImpl) SubjectFromGRPCContext(gRPCCtx context.Context) (*Subject, error) {
	tokenString, err := ah.jwtGRPCHelper.ExtractTokenStringFromContext(ah.metadataName, gRPCCtx)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract token string", err)
	}

	userInfo, err := ah.SubjectFromTokenString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract user info", err)
	}

	return userInfo, nil
}
