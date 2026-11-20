package jwt

import (
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenIDBuilderFunc определяет сигнатуру функции генерации уникальных идентификаторов токенов (JTI).
type TokenIDBuilderFunc func() (string, error)

// Factory реализует контракт TokenFactory, координируя процессы
// валидации, сборки и криптографической подписи новых JWT-токенов.
type Factory struct {
	issuer             string
	audiences          []string
	signingMethod      jwt.SigningMethod
	secretKey          []byte
	expirationInterval time.Duration
	idBuilder          TokenIDBuilderFunc
}

var _ TokenFactory = (*Factory)(nil)

// NewFactory создает и инициализирует новый экземпляр Factory для сборки и подписи токенов.
func NewFactory(
	issuer string,
	audiences []string,
	signingMethodName string,
	secretKey []byte,
	expirationInterval time.Duration,
	idBuilder TokenIDBuilderFunc,
) (*Factory, error) {
	signingMethod, err := GetSigningMethod(signingMethodName)
	if err != nil {
		return nil, err
	}

	res := &Factory{
		issuer:             issuer,
		audiences:          audiences,
		signingMethod:      signingMethod,
		secretKey:          secretKey,
		expirationInterval: expirationInterval,
		idBuilder:          idBuilder,
	}

	if utils.IsNil(idBuilder) {
		res.idBuilder = res.defaultIDBuilder
	}

	return res, nil
}

// Claims выполняет валидацию бизнес-инвариантов переданного билдера и возвращает укомплектованный объект SubjectClaims.
func (f *Factory) Claims(builder *SubjectClaimsBuilder) (*SubjectClaims, error) {
	if utils.IsNil(builder) {
		return nil, errs.NewInvalidArgumentError("builder", "is required")
	}
	claimsID, err := f.idBuilder()
	if err != nil {
		return nil, errs.NewUtlJWTError("token id builder failed", err)
	}

	// обогащаем claims необходимыми параметрами
	currentTime := time.Now()
	claims := builder.
		WithIssuer(f.issuer).
		WithAudience(f.audiences...).
		WithID(claimsID).
		WithIssuedAt(currentTime).
		WithExpiresAt(currentTime.Add(f.expirationInterval)).
		WithNotBefore(currentTime).
		Build()
	if err := claims.ValidateSubject(); err != nil {
		return nil, err
	}

	return claims, nil
}

// Token собирает неподписанный объект jwt.Token на основе подготовленных утверждений SubjectClaims.
func (f *Factory) Token(claims *SubjectClaims) (*jwt.Token, error) {
	if utils.IsNil(claims) {
		return nil, errs.NewInvalidArgumentError("claims", "is required")
	}

	return jwt.NewWithClaims(f.signingMethod, claims), nil
}

// TokenString осуществляет криптографическую подпись токена секретным ключом и возвращает финальную компактную строку JWT.
func (f *Factory) TokenString(token *jwt.Token) (string, error) {
	if utils.IsNil(token) {
		return "", errs.NewInvalidArgumentError("token", "is required")
	}

	res, err := token.SignedString(f.secretKey)
	if err != nil {
		return "", errs.NewUtlJWTError("token signing failed", err)
	}

	return res, nil
}

func (f *Factory) defaultIDBuilder() (string, error) {
	res, err := uuid.NewRandom()
	if err != nil {
		return "", errs.NewUtlJWTError("generate UUID failed", err)
	}

	return fmt.Sprintf("undef-%s", res.String()), nil
}
