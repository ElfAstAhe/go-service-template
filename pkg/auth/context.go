package auth

import (
	"context"
)

// Приватный тип ключа contextKey полностью исключает коллизии (Context Collision)
// и гарантирует, что никто снаружи пакета не сможет подделать или перезаписать Subject.
type contextKey struct{}

var (
	subjectKey = contextKey{}

	// Guest — статический иммутабельный синглтон (Null-Object pattern) для неавторизованных пользователей.
	// Имеет пустой ID, тип SubjectGuest, а мапы ролей инициализированы, что делает его 100% безопасным для чтения.
	Guest = &Subject{
		ID:    "",
		Name:  "Guest",
		Type:  SubjectGuest,
		Roles: make(map[string]struct{}),
	}
)

// WithSubject создает новый дочерний контекст на базе родительского и атомарно сохраняет туда Subject.
// Используется в транспортных Middleware/Интерцепторах после успешной криптографической верификации токена.
func WithSubject(ctx context.Context, s *Subject) context.Context {
	return context.WithValue(ctx, subjectKey, s)
}

// FromContext извлекает верифицированный объект Subject из context.Context выполнения запроса.
// ИСПРАВЛЕНО: Теперь возвращает чистый nil, если субъект отсутствует в метаданных контекста горутины,
// что позволяет вышестоящим хелперам корректно обрабатывать и логировать ошибки неавторизованных сессий.
func FromContext(ctx context.Context) *Subject {
	s, ok := ctx.Value(subjectKey).(*Subject)
	if !ok || s == nil {
		return nil
	}

	return s
}
