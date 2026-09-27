package http

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/transport"
)

// MapToHTTPStatus выполняет сквозную десериализацию и трансляцию абстрактных ошибок
// фреймворка в строго типизированные числовые статус-коды протокола HTTP (RFC 9110).
//
// Изолирует внутренние прикладные детали сбоев UseCase-слоя и DAL, гарантируя отдачу
// корректных REST-ответов клиентам и безопасность чувствительных инфраструктурных данных.
func MapToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// 400 BadRequest (Нарушение входных аргументов, инвариантов валидации или DTO-маппинга)
	if transport.IsBadRequest(err) {
		return http.StatusBadRequest
	}

	// 401 Unauthorized (Отсутствие токена, битая подпись JWT или истекший TTL сессии)
	if transport.IsUnauthorized(err) {
		return http.StatusUnauthorized
	}

	// 403 Forbidden (Недостаточно прав, провал проверки ролей в рамках RBAC контура)
	if transport.IsForbidden(err) {
		return http.StatusForbidden
	}

	// 404 NotFound (Физическое отсутствие записи в СУБД, sql.ErrNoRows)
	if transport.IsNotFound(err) {
		return http.StatusNotFound
	}

	// 409 Conflict (Нарушение уникального индекса СУБД, коллизия дублирования модели агрегата)
	if transport.IsConflict(err) {
		return http.StatusConflict
	}

	// 410 Gone (Попытка обращения к логически удаленному ресурсу в рамках паттерна Soft Delete)
	if transport.IsGone(err) {
		return http.StatusGone
	}

	// Предохранитель (Security Shield): маскируем сырые системные ошибки СУБД/Кафки под константный 500 код
	return http.StatusInternalServerError
}
