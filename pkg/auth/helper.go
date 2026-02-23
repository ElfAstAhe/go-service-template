package auth

import (
	"context"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/helper"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"
)

const (
	DefaultCookieName   string = "Authorization"
	DefaultMetadataName string = "Authorization"
)

type Helper struct {
	cookieName    string
	metadataName  string
	jwtHelper     *helper.JWTHelper
	jwtHTTPHelper *helper.JWTHTTPHelper
	jwtGRPCHelper *helper.JWTGRPCHelper
}

func NewAuthHelper(
	cookieName, metadataName string,
	jwtHelper *helper.JWTHelper,
	jwtHTTPHelper *helper.JWTHTTPHelper,
	jwtGRPCHelper *helper.JWTGRPCHelper,
) *Helper {
	return &Helper{
		cookieName:    cookieName,
		metadataName:  metadataName,
		jwtHelper:     jwtHelper,
		jwtHTTPHelper: jwtHTTPHelper,
		jwtGRPCHelper: jwtGRPCHelper,
	}
}

func NewDefaultAuthHelper(secretKey string) *Helper {
	jwtHelper := helper.NewDefaultJWTHelper(secretKey)
	jwtHTTPHelper := helper.NewJWTHTTPHelper(jwtHelper)
	jwtGRPCHelper := helper.NewJWTGRPCHelper(jwtHelper)

	return NewDefaultAuthHelperEx(jwtHelper, jwtHTTPHelper, jwtGRPCHelper)
}

func NewDefaultAuthHelperEx(
	jwtHelper *helper.JWTHelper,
	jwtHTTPHelper *helper.JWTHTTPHelper,
	jwtGRPCHelper *helper.JWTGRPCHelper,
) *Helper {
	return NewAuthHelper(DefaultCookieName, DefaultMetadataName, jwtHelper, jwtHTTPHelper, jwtGRPCHelper)
}

func (ah *Helper) SubjectFromToken(token *jwt.Token) (*Subject, error) {
	if token == nil {
		return nil, errs.NewInvalidArgumentError("token", "nil jwt token")
	}

	claims, err := ah.jwtHelper.ExtractClaims(token)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract claims", err)
	}

	return NewSubject(claims.SubjectID, claims.Subject, SubjectType(claims.SubjectType), claims.Roles, nil), nil
}

func (ah *Helper) TokenFromSubject(subject *Subject) (*jwt.Token, error) {
	if subject == nil {
		return nil, errs.NewInvalidArgumentError("subject", "nil user info")
	}

	return ah.jwtHelper.BuildToken(subject.ID, subject.Name, string(subject.Type), false)
}

func (ah *Helper) TokenStringFromSubjet(subject *Subject) (string, error) {
	token, err := ah.TokenFromSubject(subject)
	if err != nil {
		return "", err
	}

	return ah.jwtHelper.BuildTokenStr(token)
}

func (ah *Helper) SubjectFromTokenString(tokenString string) (*Subject, error) {
	token, err := ah.jwtHelper.ExtractTokenFromString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract token", err)
	}

	return ah.SubjectFromToken(token)
}

func (ah *Helper) SubjectFromContext(ctx context.Context) (*Subject, error) {
	res := FromContext(ctx)
	if res != nil {
		return res, nil
	}

	return nil, errs.NewUtlAuthError("user info not found", nil)
}

func (ah *Helper) HasSubjectInContext(ctx context.Context) bool {
	userInfo, err := ah.SubjectFromContext(ctx)
	if err != nil {
		return false
	}

	return userInfo != nil
}

func (ah *Helper) SubjectFromHTTPRequest(request *http.Request) (*Subject, error) {
	tokenString, err := ah.jwtHTTPHelper.ExtractTokenStringFromRequestCookie(ah.cookieName, request)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract token string", err)
	}

	userInfo, err := ah.SubjectFromTokenString(tokenString)
	if err != nil {
		return nil, errs.NewUtlAuthError("extract user info", err)
	}

	return userInfo, nil
}

func (ah *Helper) SubjectFromGRPCMetadata(md metadata.MD) (*Subject, error) {
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

func (ah *Helper) UserInfoFromGRPCContext(gRPCCtx context.Context) (*Subject, error) {
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
