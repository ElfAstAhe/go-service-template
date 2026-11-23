package jwt

import (
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
)

// SubjectClaims инкапсулирует полезную нагрузку (Payload) авторизационного токена.
// Расширяет стандартные W3C утверждения (RegisteredClaims) кастомными полями прав (RBAC).
type SubjectClaims struct {
	jwt.RegisteredClaims
	SubjectID   string   `json:"subject_id,omitempty"`   // Уникальный ID субъекта (например, User UUID)
	SubjectType string   `json:"subject_type,omitempty"` // Тип субъекта (user, service_account, bot)
	Roles       []string `json:"roles,omitempty"`        // Список текстовых ролей пользователя для RBAC контроля
}

var _ jwt.Claims = (*SubjectClaims)(nil)

// newSubjectClaims создает пустой экземпляр SubjectClaims для десериализации токена.
func newSubjectClaims() *SubjectClaims {
	return &SubjectClaims{}
}

// ValidateSubject проверяет наличие обязательных бизнес-полей идентичности субъекта в токене.
func (sc *SubjectClaims) ValidateSubject() error {
	if strings.TrimSpace(sc.Subject) == "" {
		return errs.NewUtlJWTError("subject required", nil)
	}
	if strings.TrimSpace(sc.SubjectType) == "" {
		return errs.NewUtlJWTError("subject type required", nil)
	}
	//if strings.TrimSpace(sc.SubjectID) == "" {
	//	return errs.NewUtlJWTError("subject id required", nil)
	//}

	return nil
}
