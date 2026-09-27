package grpc

import (
	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapToGrpcError выполняет сквозную десериализацию и трансляцию абстрактных ошибок
// фреймворка в строго типизированные сетевые gRPC-статусы (gRPC Status Error Codes).
//
// Изолирует внутренние прикладные детали сбоев UseCase-слоя и DAL, гарантируя отдачу
// корректных кодов ответов клиентам и безопасность инфраструктурных данных.
func MapToGrpcError(err error) error {
	if err == nil {
		return nil
	}

	// Bad Request (Нарушение входных аргументов, инвариантов валидации или DTO-маппинга)
	if transport.IsBadRequest(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Unauthorized (Отсутствие токена, битая подпись JWT или истекший TTL)
	if transport.IsUnauthorized(err) {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	// Forbidden (Недостаточно прав, провал проверки ролей в рамках RBAC контура)
	if transport.IsForbidden(err) {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	// Not Found (Физическое отсутствие записи в СУБД, sql.ErrNoRows)
	if transport.IsNotFound(err) {
		return status.Error(codes.NotFound, err.Error())
	}

	// Conflict (Нарушение уникального индекса СУБД, коллизия дублирования модели)
	if transport.IsConflict(err) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	// Gone (Попытка обращения к логически удаленному ресурсу в рамках паттерна Soft Delete)
	// 💡 Поскольку gRPC спецификация не содержит аналога HTTP 410 Gone, маппим в каноничный codes.NotFound.
	if transport.IsGone(err) {
		return status.Error(codes.NotFound, err.Error())
	}

	// Предохранитель (Security Shield): маскируем сырые системные ошибки СУБД/Кафки под константный Internal код
	return status.Error(codes.Internal, "internal server error")
}
