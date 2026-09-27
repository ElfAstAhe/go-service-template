package helper

import (
	"errors"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Рантайм-дефолты для безопасной конфигурации JWT-токенов в экосистеме.
const (
	DefaultJWTSigningMethodName  string        = "HS256"               // Метод подписи по умолчанию: HMAC-SHA256
	DefaultJWTExpirationDuration time.Duration = 15 * time.Minute      // Стандартное время жизни Access-токена (15 минут)
	DefaultJWTIssuer             string        = "go-service-template" // Идентификатор издателя токена (Issuer)
)

const (
	TokenPrefix string = "Bearer " // Стандартный префикс HTTP-заголовка Authorization
)

// Глобальный синглтон дефолтного метода подписи, извлеченный из реестра golang-jwt.
var (
	DefaultJWTSigningMethod = jwt.GetSigningMethod(DefaultJWTSigningMethodName)
)

// TokenIDBuilder описывает сигнатуру функции-генератора уникальных идентификаторов токенов (JTI).
type TokenIDBuilder func() string

// AppClaims инкапсулирует полезную нагрузку (Payload) авторизационного токена.
// Расширяет стандартные W3C утверждения (RegisteredClaims) кастомными полями прав (RBAC).
type AppClaims struct {
	*jwt.RegisteredClaims
	Admin       bool     `json:"admin,omitempty"`        // Флаг суперпользователя (root)
	SubjectID   string   `json:"subject_id,omitempty"`   // Уникальный ID субъекта (например, User UUID)
	SubjectType string   `json:"subject_type,omitempty"` // Тип субъекта (user, service_account, bot)
	Roles       []string `json:"roles,omitempty"`        // Список текстовых ролей пользователя для RBAC контроля
}

// NewAppClaims — конструктор payload JWT токена со строгим заполнением таймштампов выпуска и экспирации.
func NewAppClaims(
	subjectID string,
	subject string,
	subjectType string,
	admin bool,
	tokenIDBuilder TokenIDBuilder,
	issuer string,
	expirationDuration time.Duration,
	roles ...string,
) *AppClaims {
	return &AppClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			ID:        tokenIDBuilder(),                                       // Уникальный JTI токена
			Issuer:    issuer,                                                 // Кем выдан
			IssuedAt:  jwt.NewNumericDate(time.Now()),                         // Время выдачи
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationDuration)), // Время смерти (TTL)
			Subject:   subject,                                                // Логин/Email
		},
		Admin:       admin,
		SubjectID:   subjectID,
		SubjectType: subjectType,
		Roles:       roles,
	}
}

// NewEmptyAppClaims генерирует пустую заготовку AppClaims, необходимую парсеру для десериализации JSON.
func NewEmptyAppClaims() *AppClaims {
	return &AppClaims{
		RegisteredClaims: &jwt.RegisteredClaims{},
		Roles:            make([]string, 0),
	}
}

// JWTHelper координирует процессы выпуска, подписи, парсинга и криптографической верификации JWT-токенов.
type JWTHelper struct {
	issuer             string            // Издатель текущего инстанса
	signingMethod      jwt.SigningMethod // Алгоритм криптографической подписи (HMAC / RSA / ECDSA)
	secretKey          string            // Секретная соль/ключ для валидации и генерации подписей
	expirationDuration time.Duration     // Лимит времени жизни токенов текущего хелпера
	tokenIDBuilder     TokenIDBuilder    // Инжектированный генератор JTI
}

// NewJWTHelper — фабричный конструктор менеджера JWT.
func NewJWTHelper(
	issuer string,
	signingMethod jwt.SigningMethod,
	secretKey string,
	expirationDuration time.Duration,
	tokenIDBuilder TokenIDBuilder,
) *JWTHelper {
	return &JWTHelper{
		issuer:             issuer,
		signingMethod:      signingMethod,
		secretKey:          secretKey,
		expirationDuration: expirationDuration,
		tokenIDBuilder:     tokenIDBuilder,
	}
}

// NewDefaultJWTHelper собирает менеджер JWT, наполняя его системными константными дефолтами.
func NewDefaultJWTHelper(secretKey string) *JWTHelper {
	return NewJWTHelper(DefaultJWTIssuer, DefaultJWTSigningMethod, secretKey, DefaultJWTExpirationDuration, defaultTokenIDBuilder)
}

// ExtractClaims извлекает и типизирует доменные утверждения AppClaims из верифицированного токена.
func (h *JWTHelper) ExtractClaims(token *jwt.Token) (*AppClaims, error) {
	res, ok := token.Claims.(*AppClaims)
	if !ok {
		return nil, errs.NewUtlJWTError("invalid claims format", nil)
	}

	return res, nil
}

// ExtractTokenFromString осуществляет парсинг сырой текстовой строки токена,
// проверяет математическую валидность цифровой подписи и осуществляет защиту от атак перегрузки алгоритма.
func (h *JWTHelper) ExtractTokenFromString(tokenString string) (*jwt.Token, error) {
	claims := NewEmptyAppClaims()
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// 🛡️ Защитный барьер: проверяем совпадение алгоритмов во избежание Algorithm Confusion атак
		if h.signingMethod.Alg() != token.Method.Alg() {
			return nil, errs.NewUtlJWTError("invalid signing method algorithm", nil)
		}

		return []byte(h.secretKey), nil
	})

	if err != nil {
		// Используем современный дженерик-распаковщик ошибок Go 1.23+ для проброса исходной ошибки
		if _, ok := errors.AsType[*errs.UtlJWTError](err); !ok {
			return nil, errs.NewUtlJWTError("parse jwt failed", err)
		}

		return nil, err
	}

	// Финальная рантайм верификация валидности и таймштампов токена (Expired / Not Before)
	if !token.Valid {
		return nil, errs.NewUtlJWTError("token validation failed / token is invalid", nil)
	}

	return token, nil
}

// BuildClaims генерирует укомплектованный валидный объект AppClaims с валидацией обязательных бизнес-полей.
func (h *JWTHelper) BuildClaims(subjectID, subject, subjectType string, admin bool, roles ...string) (*AppClaims, error) {
	if subjectID == "" {
		return nil, errs.NewInvalidArgumentError("subject ID", "subject ID is empty")
	}
	if subject == "" {
		return nil, errs.NewInvalidArgumentError("subject", "subject is empty")
	}

	return NewAppClaims(subjectID, subject, subjectType, admin, h.buildTokenID, h.issuer, h.expirationDuration, roles...), nil
}

// BuildToken собирает неподписанный объект jwt.Token на основе сгенерированных claims.
func (h *JWTHelper) BuildToken(subjectID, subject, subjectType string, admin bool, roles ...string) (*jwt.Token, error) {
	claims, err := h.BuildClaims(subjectID, subject, subjectType, admin, roles...)
	if err != nil {
		return nil, errs.NewUtlJWTError("build claims failed", err)
	}

	return jwt.NewWithClaims(h.signingMethod, claims), nil
}

// BuildTokenStr подписывает готовый токен секретным ключом, превращая его в финальную компактную строку.
func (h *JWTHelper) BuildTokenStr(token *jwt.Token) (string, error) {
	if token == nil {
		return "", errs.NewUtlJWTError("nil token reference", nil)
	}

	res, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		return "", errs.NewUtlJWTError("sign token failed", err)
	}

	return res, nil
}

// BuildTokenString — сквозной метод полного цикла выпуска токена: от генерации до подписи и выдачи строки.
func (h *JWTHelper) BuildTokenString(subjectID, subject, subjectType string, admin bool, roles ...string) (string, error) {
	token, err := h.BuildToken(subjectID, subject, subjectType, admin, roles...)
	if err != nil {
		return "", errs.NewUtlJWTError("building token failed", err)
	}

	return h.BuildTokenStr(token)
}

// buildTokenID маршрутизирует вызов на инжектированный или дефолтный генератор JTI.
func (h *JWTHelper) buildTokenID() string {
	if h.tokenIDBuilder != nil {
		return h.tokenIDBuilder()
	}

	return defaultTokenIDBuilder()
}

// defaultTokenIDBuilder генерирует префиксированный случайный UUID для JTI-меток токенов.
func defaultTokenIDBuilder() string {
	template := "undef-%v"
	res, err := uuid.NewRandom()
	if err != nil {
		// Страховочный Fallback на наносекунды в случае сбоя криптографической энтропии операционной системы
		return fmt.Sprintf(template, time.Now().Nanosecond())
	}

	return fmt.Sprintf(template, res.String())
}
