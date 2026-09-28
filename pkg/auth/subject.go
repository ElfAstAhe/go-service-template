package auth

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// SubjectType определяет строго типизированное строковое перечисление типов субъектов авторизации (Client/Subject Categories).
type SubjectType string

// Набор легитимных констант категорий клиентов в системе.
const (
	SubjectUser    SubjectType = "user"    // Физический пользователь-человек (End-User)
	SubjectService SubjectType = "service" // Сервисный аккаунт / Внутренний межсервисный клиент (M2M / Service Account)
	SubjectGuest   SubjectType = "guest"   // Неаутентифицированный анонимный посетитель (Anonymous/Guest)
)

// Глобальный неизменяемый синглтон-реестр допустимых категорий для O(1) валидации.
var allowedSubjectTypes = map[SubjectType]struct{}{
	SubjectUser:    {},
	SubjectService: {},
	SubjectGuest:   {},
}

// IsValid проверяет за константное время O(1), входит ли текущий тип в пул легитимных категорий платформы.
//
//goland:noinspection GoMixedReceiverTypes
func (st SubjectType) IsValid() bool {
	_, ok := allowedSubjectTypes[st]
	return ok
}

// ParseSubjectType преобразует сырую строку в строго валидированный доменный тип SubjectType.
func ParseSubjectType(str string) (SubjectType, error) {
	var res = SubjectType(str)
	if !res.IsValid() {
		return "", errs.NewInvalidArgumentError("str", fmt.Sprintf("subject type [%s] not allowed", str))
	}

	return res, nil
}

// UnmarshalText реализует интерфейс encoding.TextUnmarshaler для автоматической валидации типа при JSON/YAML десериализации.
//
//goland:noinspection GoMixedReceiverTypes
func (st *SubjectType) UnmarshalText(text []byte) error {
	val := SubjectType(text)
	if !val.IsValid() {
		return errs.NewInvalidArgumentError("text", fmt.Sprintf("invalid subject type: %s", string(text)))
	}
	*st = val
	return nil
}

// MarshalText реализует интерфейс encoding.TextMarshaler для каноничной текстовой сериализации.
func (st SubjectType) MarshalText() ([]byte, error) {
	return []byte(st), nil
}

// Subject представляет из себя аутентифицированного и авторизованного юзверя (Security Context Identity).
//
// Инкапсулирует его первичные идентификаторы, категорию клиента, оптимизированную карту ролей для контроля
// политик доступа (@RolesAllowed) и метаданные сетевого окружения.
type Subject struct {
	ID       string              // Уникальный идентификатор субъекта (соответствует "sub" в спецификации W3C JWT)
	Name     string              // Печатное имя или логин/email пользователя
	Type     SubjectType         // Критериальный тип/категория клиента (user/service/guest)
	Roles    map[string]struct{} // Карта-множество ролей, оптимизированная для сверхбыстрой O(1) проверки политик RBAC
	Metadata map[string]string   // Дополнительный контекст безопасности рантайма (IP-адрес, User-Agent, DeviceID)
}

// NewSubject — фабричный конструктор доменной модели субъекта безопасности с изоляцией срезов (Deep Copy).
func NewSubject(id, name string, subjectType SubjectType, roles []string, metadata map[string]string) *Subject {
	// Преобразуем плоский срез ролей в высокопроизводительное хэш-множество
	mapRoles := make(map[string]struct{})
	for _, role := range roles {
		mapRoles[role] = struct{}{}
	}

	// Выполняем глубокое копирование метаданных во избежание сайд-эффектов разделяемой памяти
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

// HasRole проверяет наличие у субъекта запрашиваемой роли за O(1) (аналог JAAS securityContext.isCallerInRole).
func (s *Subject) HasRole(role string) bool {
	_, ok := s.Roles[role]
	return ok
}

// IsAuthenticated проверяет, прошел ли данный субъект криптографическую проверку подлинности (что он не Anonymous).
func (s *Subject) IsAuthenticated() bool {
	return s.Type != SubjectGuest && s.ID != ""
}

// IsUser возвращает истину, если субъект является физическим пользователем.
func (s *Subject) IsUser() bool {
	return s.Type == SubjectUser
}

// IsService возвращает истину, если субъект является внутренним системным аккаунтом микросервиса.
func (s *Subject) IsService() bool {
	return s.Type == SubjectService
}

// IsGuest возвращает истину, если субъект является неавторизованным гостем.
func (s *Subject) IsGuest() bool {
	return s.Type == SubjectGuest
}

// IsValid осуществляет базовую проверку структурной заполненности полей идентичности.
func (s *Subject) IsValid() bool {
	return s.ID != "" && s.Name != "" && s.Type != ""
}

// String формирует компактный текстовый паспорт идентичности субъекта для сквозного логгирования UseCase-операций.
func (s *Subject) String() string {
	return s.ID + "@" + s.Name
}
