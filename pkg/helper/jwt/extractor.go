package jwt

import (
	"fmt"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

// Extractor реализует контракт TokenExtractor, разделяя криптографическую проверку подписи токена
// и валидацию стандартных и доменных утверждений SubjectClaims.
type Extractor struct {
	claimsValidator *jwt.Validator
	tokenParseOpts  []jwt.ParserOption
	secretKey       []byte
}

var _ TokenExtractor = (*Extractor)(nil)

// NewExtractor создает новый экземпляр Extractor с заданными правилами проверки подписи и утверждений JWT.
func NewExtractor(
	secretKey []byte,
	signingMethodName string,
	audience string,
	issuer string,
	leeway time.Duration,
) (*Extractor, error) {
	if len(secretKey) == 0 {
		return nil, errs.NewInvalidArgumentError("secretKey", "is required")
	}
	if _, err := GetSigningMethod(signingMethodName); err != nil {
		return nil, err
	}

	// claims validator
	claimsValidator := jwt.NewValidator(
		jwt.WithAudience(audience),
		jwt.WithIssuer(issuer),
		jwt.WithIssuedAt(),
		jwt.WithNotBeforeRequired(),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(leeway),
	)
	// token parse options
	tokenParseOpts := []jwt.ParserOption{
		jwt.WithoutClaimsValidation(),
		jwt.WithValidMethods([]string{signingMethodName}),
	}

	// instance
	return &Extractor{
		claimsValidator: claimsValidator,
		tokenParseOpts:  tokenParseOpts,
		secretKey:       secretKey,
	}, nil
}

// Claims извлекает утверждения SubjectClaims из токена и проверяет стандартные (aud, iss, exp, nbf, iat)
// и доменные (subject, subject_type, subject_id) инварианты.
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
	if err := e.claimsValidator.Validate(claims); err != nil {
		return nil, errs.NewUtlJWTError("claims validation failed", err)
	}
	if err := claims.ValidateSubject(); err != nil {
		return nil, errs.NewUtlJWTError("claims subject validation failed", err)
	}

	return claims, nil
}

// Token парсит строку JWT и проверяет алгоритм и цифровую подпись без валидации утверждений.
// Для полной проверки сроков действия и полей субъекта полученный токен необходимо передать в Claims.
func (e *Extractor) Token(tokenStr string) (*jwt.Token, error) {
	if strings.TrimSpace(tokenStr) == "" {
		return nil, errs.NewInvalidArgumentError("token", "is required")
	}
	token, err := jwt.ParseWithClaims(tokenStr, newSubjectClaims(), func(_ *jwt.Token) (any, error) {
		return e.secretKey, nil
	}, e.tokenParseOpts...)
	if err != nil {
		return nil, errs.NewUtlJWTError("token parse failed", err)
	}

	return token, nil
}
