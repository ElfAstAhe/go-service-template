package auth

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type SubjectType string

const (
	SubjectUser    SubjectType = "user"
	SubjectService SubjectType = "service"
	SubjectGuest   SubjectType = "guest"
)

var allowedSubjectTypes = map[SubjectType]struct{}{
	SubjectUser:    {},
	SubjectService: {},
	SubjectGuest:   {},
}

func (st SubjectType) IsValid() bool {
	_, ok := allowedSubjectTypes[st]
	return ok
}

func ParseSubjectType(str string) (SubjectType, error) {
	var res = SubjectType(str)
	if !res.IsValid() {
		return "", errs.NewInvalidArgumentError("str", fmt.Sprintf("subject type [%s] not allowed", str))
	}

	return res, nil
}

func (st *SubjectType) UnmarshalText(text []byte) error {
	val := SubjectType(text)
	if !val.IsValid() {
		return errs.NewInvalidArgumentError("text", fmt.Sprintf("invalid subject type: %s", string(text)))
	}
	*st = val
	return nil
}

func (st SubjectType) MarshalText() ([]byte, error) {
	return []byte(st), nil
}

// Subject представляет из себя аутентифицированного и авторизованного юзверя
type Subject struct {
	ID       string // sub в JWT
	Name     string
	Type     SubjectType         // тип клиента
	Roles    map[string]struct{} // Оптимизировано для быстрой проверки @RolesAllowed
	Metadata map[string]string   // Доп. данные (IP, DeviceID)
}

func NewSubject(id, name string, subjectType SubjectType, roles []string, metadata map[string]string) *Subject {
	mapRoles := make(map[string]struct{})
	for _, role := range roles {
		mapRoles[role] = struct{}{}
	}
	mapMetadata := make(map[string]string, len(metadata))
	for k, v := range metadata {
		mapMetadata[k] = v
	}

	return &Subject{
		ID:       id,
		Name:     name,
		Type:     subjectType,
		Roles:    mapRoles,
		Metadata: mapMetadata,
	}
}

// HasRole — аналог securityContext.isCallerInRole(role)
func (s *Subject) HasRole(role string) bool {
	_, ok := s.Roles[role]
	return ok
}

// IsAuthenticated — проверка, что это не Anonymous
func (s *Subject) IsAuthenticated() bool {
	return s.Type != SubjectGuest && s.ID != ""
}

func (s *Subject) IsUser() bool {
	return s.Type == SubjectUser
}

func (s *Subject) IsService() bool {
	return s.Type == SubjectService
}

func (s *Subject) IsGuest() bool {
	return s.Type == SubjectGuest
}

func (s *Subject) IsValid() bool {
	return s.ID != "" && s.Name != "" && s.Type != ""
}

func (s *Subject) String() string {
	return s.ID + "@" + s.Name
}
