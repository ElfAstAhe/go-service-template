package jwt

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

// SubjectClaimsBuilder реализует паттерн Fluent Builder для поэтапного
// и изолированного конструирования объектов SubjectClaims.
type SubjectClaimsBuilder struct {
	instance *SubjectClaims
}

var _ utils.Builder[*SubjectClaims] = (*SubjectClaimsBuilder)(nil)

// NewSubjectClaimsBuilder создает и инициализирует новый экземпляр билдера.
func NewSubjectClaimsBuilder() *SubjectClaimsBuilder {
	return &SubjectClaimsBuilder{
		instance: newSubjectClaims(),
	}
}

// WithIssuer устанавливает издателя (iss) токена.
func (scb *SubjectClaimsBuilder) WithIssuer(issuer string) *SubjectClaimsBuilder {
	scb.instance.Issuer = issuer
	return scb
}

// WithSubject устанавливает человекочитаемое имя субъекта (sub).
func (scb *SubjectClaimsBuilder) WithSubject(subject string) *SubjectClaimsBuilder {
	scb.instance.Subject = subject
	return scb
}

// WithAudience устанавливает список целевых аудиторий (aud) токена.
func (scb *SubjectClaimsBuilder) WithAudience(audience ...string) *SubjectClaimsBuilder {
	scb.instance.Audience = audience
	return scb
}

// WithExpiresAt устанавливает время окончания действия токена (exp).
func (scb *SubjectClaimsBuilder) WithExpiresAt(expiresAt time.Time) *SubjectClaimsBuilder {
	scb.instance.ExpiresAt = nil
	if !expiresAt.IsZero() {
		scb.instance.ExpiresAt = jwt.NewNumericDate(expiresAt)
	}
	return scb
}

// WithNotBefore устанавливает время, начиная с которого токен становится валидным (nbf).
func (scb *SubjectClaimsBuilder) WithNotBefore(before time.Time) *SubjectClaimsBuilder {
	scb.instance.NotBefore = nil
	if !before.IsZero() {
		scb.instance.NotBefore = jwt.NewNumericDate(before)
	}
	return scb
}

// WithIssuedAt устанавливает время выпуска токена (iat).
func (scb *SubjectClaimsBuilder) WithIssuedAt(issuedAt time.Time) *SubjectClaimsBuilder {
	scb.instance.IssuedAt = nil
	if !issuedAt.IsZero() {
		scb.instance.IssuedAt = jwt.NewNumericDate(issuedAt)
	}
	return scb
}

// WithID устанавливает уникальный идентификатор токена (jti).
func (scb *SubjectClaimsBuilder) WithID(id string) *SubjectClaimsBuilder {
	scb.instance.ID = id
	return scb
}

// WithSubjectID устанавливает уникальный доменный ID субъекта (UUID).
func (scb *SubjectClaimsBuilder) WithSubjectID(subjectID string) *SubjectClaimsBuilder {
	scb.instance.SubjectID = subjectID
	return scb
}

// WithSubjectType устанавливает тип субъекта (user, bot, service_account).
func (scb *SubjectClaimsBuilder) WithSubjectType(subjectType string) *SubjectClaimsBuilder {
	scb.instance.SubjectType = subjectType
	return scb
}

// WithAdmin устанавливает флаг суперпользователя.
func (scb *SubjectClaimsBuilder) WithAdmin(flag bool) *SubjectClaimsBuilder {
	scb.instance.Admin = flag
	return scb
}

// WithRoles заполняет список RBAC-ролей субъекта.
func (scb *SubjectClaimsBuilder) WithRoles(roles ...string) *SubjectClaimsBuilder {
	scb.instance.Roles = roles
	return scb
}

// Build завершает конструирование, возвращает сформированный объект *SubjectClaims
// и атомарно сбрасывает состояние билдера для безопасного повторного использования.
func (scb *SubjectClaimsBuilder) Build() *SubjectClaims {
	res := scb.instance
	scb.instance = newSubjectClaims()
	return res
}
