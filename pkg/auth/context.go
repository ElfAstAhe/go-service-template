package auth

import (
	"context"
)

// Приватный тип ключа — никто снаружи не сможет подделать или затереть Subject
type contextKey struct{}

var (
	subjectKey = contextKey{}

	// Guest — статический объект для неавторизованных пользователей.
	// ID пустой, тип Guest, мапы инициализированы (безопасно для чтения).
	Guest = &Subject{
		ID:    "",
		Name:  "Guest",
		Type:  SubjectGuest,
		Roles: make(map[string]struct{}),
	}
)

// WithSubject создает новый контекст на базе родительского и кладет туда Subject.
// Используется в Middleware после успешной аутентификации.
func WithSubject(ctx context.Context, s *Subject) context.Context {
	return context.WithValue(ctx, subjectKey, s)
}

// FromContext извлекает Subject из контекста.
// Если Subject не был установлен (например, забыли Middleware),
// возвращает объект Guest, чтобы методы .HasRole() не паниковали.
func FromContext(ctx context.Context) *Subject {
	s, ok := ctx.Value(subjectKey).(*Subject)
	if !ok || s == nil {
		return Guest
	}

	return s
}
