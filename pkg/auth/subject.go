package auth

type SubjectType string

const (
	SubjectUser    SubjectType = "user"
	SubjectService SubjectType = "service"
	SubjectGuest   SubjectType = "guest"
)

type Subject struct {
	ID       string              // sub в JWT
	Type     SubjectType         // тип клиента
	Roles    map[string]struct{} // Оптимизировано для быстрой проверки @RolesAllowed
	Metadata map[string]string   // Доп. данные (IP, DeviceID)
}

// HasRole — аналог securityContext.isCallerInRole(role)
func (s Subject) HasRole(role string) bool {
	_, ok := s.Roles[role]
	return ok
}

// IsAuthenticated — проверка, что это не Anonymous
func (s Subject) IsAuthenticated() bool {
	return s.Type != SubjectGuest && s.ID != ""
}
