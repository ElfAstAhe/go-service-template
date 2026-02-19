package auth

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

func (st SubjectType) FromString(str string) SubjectType {

}

type Subject struct {
	ID       string              // sub в JWT
	Type     SubjectType         // тип клиента
	Roles    map[string]struct{} // Оптимизировано для быстрой проверки @RolesAllowed
	Metadata map[string]string   // Доп. данные (IP, DeviceID)
}

func NewSubject(id, subjectType, roles []string, metadata map[string]string) *Subject {
	return &Subject{}
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
