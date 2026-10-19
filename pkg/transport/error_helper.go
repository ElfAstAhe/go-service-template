package transport

import (
	"errors"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// IsBadRequest классифицирует ошибки, вызванные некорректными или поврежденными входными данными от клиента.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 400 (Bad Request)
//   - В gRPC: статус-код codes.InvalidArgument (3)
func IsBadRequest(err error) bool {
	_, okInvalidArgument := errors.AsType[*errs.InvalidArgumentError](err) // Ошибка валидации входящего DTO фреймворком
	_, okBllValidate := errors.AsType[*errs.BllValidateError](err)         // Ошибка нарушения бизнес-инвариантов на уровне домена
	_, okTlMapping := errors.AsType[*errs.TlMappingError](err)             // Ошибка маппинга структур на транспортном слое

	return okInvalidArgument || okBllValidate || okTlMapping
}

// IsUnauthorized проверяет, вызвана ли ошибка отсутствием, истечением или невалидностью токена/сессии.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 401 (Unauthorized)
//   - В gRPC: статус-код codes.Unauthenticated (16)
func IsUnauthorized(err error) bool {
	_, okBllUnauthorized := errors.AsType[*errs.BllUnauthorizedError](err) // Ошибка аутентификации (протухший JWT, неверный пароль/API-ключ)

	return okBllUnauthorized
}

// IsForbidden проверяет ошибки, связанные с нехваткой прав, ролей или скоупов (RBAC/ABAC) у субъекта запроса.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 403 (Forbidden)
//   - В gRPC: статус-код codes.PermissionDenied (7)
func IsForbidden(err error) bool {
	_, okBllForbidden := errors.AsType[*errs.BllForbiddenError](err) // Личность опознана, но у текущего пользователя нет прав на этот ресурс/действие

	return okBllForbidden
}

// IsNotFound проверяет, вызвана ли ошибка физическим или логическим отсутствием запрашиваемого ресурса в системе.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 404 (Not Found)
//   - В gRPC: статус-код codes.NotFound (5)
func IsNotFound(err error) bool {
	_, okBllNotFound := errors.AsType[*errs.BllNotFoundError](err) // Логическое отсутствие объекта на уровне бизнес-сценария (указан неверный UUID)
	_, okDalNotFound := errors.AsType[*errs.DalNotFoundError](err) // Физическое отсутствие строки в базе данных (sql.ErrNoRows на уровне Helper)

	return okBllNotFound || okDalNotFound
}

// IsConflict выявляет ошибки состояния системы, когда операция нарушает консистентность или уникальность данных.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 409 (Conflict)
//   - В gRPC: статус-код codes.AlreadyExists (6) или codes.Aborted (10)
func IsConflict(err error) bool {
	_, okBllUnique := errors.AsType[*errs.BllUniqueError](err)               // Конфликт уникальности, пойманный на уровне бизнес-логики домена
	_, okDalAlreadyExists := errors.AsType[*errs.DalAlreadyExistsError](err) // Нарушение уникального индекса (Unique Violation) на уровне СУБД (PostgreSQL)

	return okBllUnique || okDalAlreadyExists
}

// IsGone проверяет, удален ли запрашиваемый ресурс из системы окончательно и безвозвратно.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 410 (Gone)
//   - В gRPC: статус-код codes.NotFound (5) или codes.FailedPrecondition (9)
func IsGone(err error) bool {
	_, okDalSoftDeleted := errors.AsType[*errs.DalSoftDeletedError](err) // Ресурс помечен как удаленный (Soft Delete) в DAL-слое базы данных

	return okDalSoftDeleted
}
