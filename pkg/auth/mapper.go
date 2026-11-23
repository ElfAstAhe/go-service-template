package auth

import (
	"maps"
	"slices"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/helper/jwt"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// MapClaimsToSubject преобразует утверждения JWT-токена SubjectClaims в доменную модель авторизации Subject.
func MapClaimsToSubject(claims *jwt.SubjectClaims) (*Subject, error) {
	if utils.IsNil(claims) {
		return nil, errs.NewInvalidArgumentError("claims", "is required")
	}

	subjectType, err := ParseSubjectType(claims.SubjectType)
	if err != nil {
		return nil, errs.NewCommonError("subject type conversion failed", err)
	}

	return NewSubject(
		claims.SubjectID,
		claims.Subject,
		subjectType,
		claims.Roles,
		nil,
	), nil
}

// MapSubjectToClaimsBuilder заполняет SubjectClaimsBuilder базовыми полями идентичности и ролей из доменного Subject.
func MapSubjectToClaimsBuilder(subject *Subject) (*jwt.SubjectClaimsBuilder, error) {
	if utils.IsNil(subject) {
		return nil, errs.NewInvalidArgumentError("subject", "is required")
	}

	return jwt.NewSubjectClaimsBuilder().
		WithSubject(subject.Name).
		WithSubjectID(subject.ID).
		WithSubjectType(string(subject.Type)).
		WithRoles(slices.Collect(maps.Keys(subject.Roles))...), nil
}
