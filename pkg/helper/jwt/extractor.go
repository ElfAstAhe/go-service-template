package jwt

import (
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

type Extractor struct {
	validator *jwt.Validator
}

var _ TokenExtractor = (*Extractor)(nil)

func NewExtractor(
	signingMethodName string,
	audience string,
	issuer string,
	leeway time.Duration,
) *Extractor {
	return &Extractor{
		validator: jwt.NewValidator(
			jwt.WithAudience(audience),
			jwt.WithValidMethods([]string{signingMethodName}),
			jwt.WithExpirationRequired(),
			jwt.WithIssuer(issuer),
			jwt.WithIssuedAt(),
			jwt.WithNotBeforeRequired(),
			jwt.WithExpirationRequired(),
			jwt.WithLeeway(leeway),
		),
	}
}

func (e *Extractor) Claims(token *jwt.Token) (*SubjectClaims, error) {
	if utils.IsNil(token) {
		return nil, errs.NewInvalidArgumentError("token", "is required")
	}
	if utils.IsNil(token.Claims) {
		return nil, errs.NewUtlJWTError("token claims is required", nil)
	}

	claims, ok := token.Claims.(*SubjectClaims)
	if !ok {
		return nil, errs.NewUtlJWTError(fmt.Sprintf("unsupported token claims type [%s]", utils.GetFullTypeName(token.Claims)), nil)
	}

	return claims, nil
}

func (e *Extractor) Token(tokenStr string) (*jwt.Token, error) {
	//TODO implement me
	panic("implement me")
}
