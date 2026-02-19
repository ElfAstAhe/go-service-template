package helper

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	DefaultJWTSigningMethodName  string = "HS256"
	DefaultJWTExpirationDuration        = 30 * time.Minute
	DefaultJWTIssuer             string = "goph-keeper"
)

const (
	TokenPrefix string = "Bearer "
)

var (
	DefaultJWTSigningMethod = jwt.GetSigningMethod(DefaultJWTSigningMethodName)
)

type TokenIDBuilder func() string

type AppClaims struct {
	jwt.RegisteredClaims
	Admin  bool     `json:"admin,omitempty"`
	UserID string   `json:"user_id,omitempty"`
	Roles  []string `json:"roles,omitempty"`
}

func NewAppClaims(
	userID string,
	user string,
	admin bool,
	tokenIDBuilder TokenIDBuilder,
	issuer string,
	expirationDuration time.Duration,
	roles ...string,
) *AppClaims {
	return &AppClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenIDBuilder(),
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationDuration)),
			Subject:   user,
		},
		Admin:  admin,
		UserID: userID,
		Roles:  roles,
	}
}

type JWTHelper struct {
	signingMethod      jwt.SigningMethod
	secretKey          string
	expirationDuration time.Duration
	tokenIDBuilder     TokenIDBuilder
}

func NewJWTHelper(signingMethod jwt.SigningMethod, secretKey string, expirationDuration time.Duration, tokenIDBuilder TokenIDBuilder) *JWTHelper {
	return &JWTHelper{
		signingMethod:      signingMethod,
		secretKey:          secretKey,
		expirationDuration: expirationDuration,
		tokenIDBuilder:     tokenIDBuilder,
	}
}

func NewDefaultJWTHelper(secretKey string) *JWTHelper {
	return NewJWTHelper(DefaultJWTSigningMethod, secretKey, DefaultJWTExpirationDuration, defaultTokenIDBuilder)
}

func (h *JWTHelper) ExtractClaims(token *jwt.Token) (*AppClaims, error) {
	res, ok := token.Claims.(*AppClaims)
	if !ok {
		return nil, errs.NewUtlJWTError("invalid claims", nil)
	}

	return res, nil
}

func (h *JWTHelper) ExtractTokenFromString(tokenString string) (*jwt.Token, error) {
	claims := new(AppClaims)
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if reflect.TypeOf(h.signingMethod) != reflect.TypeOf(token.Method) {
			return nil, errs.NewUtlJWTError("invalid signing method", nil)
		}

		return []byte(h.secretKey), nil
	})

	if err != nil {
		if !errors.As(err, &errs.ErrUtlJWT) {
			return nil, errs.NewUtlJWTError("error parse jwt", err)
		}

		return nil, err
	}

	if !token.Valid {
		return nil, errs.NewUtlJWTError("token validation failed", nil)
	}

	return token, nil
}

func (h *JWTHelper) BuildClaims(userID, username string, admin bool, roles ...string) (*AppClaims, error) {
	if userID == "" {
		return nil, errs.NewInvalidArgumentError("user ID", "user ID is empty")
	}
	if username == "" {
		return nil, errs.NewInvalidArgumentError("user name", "user name is empty")
	}

	return NewAppClaims(userID, username, admin, h.buildTokenID, DefaultJWTIssuer, h.expirationDuration, roles...), nil
}

func (h *JWTHelper) BuildToken(userID, username string, admin bool, roles ...string) (*jwt.Token, error) {
	claims, err := h.BuildClaims(userID, username, admin, roles...)
	if err != nil {
		return nil, errs.NewUtlJWTError("error building claims", err)
	}

	return jwt.NewWithClaims(h.signingMethod, claims), nil
}

func (h *JWTHelper) BuildTokenStr(token *jwt.Token) (string, error) {
	if token == nil {
		return "", errs.NewUtlJWTError("nil token", nil)
	}

	res, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		return "", errs.NewUtlJWTError("error signing token", err)
	}

	return res, nil
}

func (h *JWTHelper) BuildTokenString(userID, username string, admin bool, roles ...string) (string, error) {
	token, err := h.BuildToken(userID, username, admin, roles...)
	if err != nil {
		return "", errs.NewUtlJWTError("error building token", err)
	}

	return h.BuildTokenStr(token)
}

func (h *JWTHelper) buildTokenID() string {
	if h.tokenIDBuilder != nil {
		return h.tokenIDBuilder()
	}

	return defaultTokenIDBuilder()
}

func defaultTokenIDBuilder() string {
	template := "undef-%v"
	res, err := uuid.NewRandom()
	if err != nil {
		return fmt.Sprintf(template, time.Now().Nanosecond())
	}

	return fmt.Sprintf(template, res.String())
}
