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
	var (
		errInvalidArgument *errs.InvalidArgumentError // Ошибка валидации входящего DTO фреймворком
		errBllValidate     *errs.BllValidateError     // Ошибка нарушения бизнес-инвариантов на уровне домена
		errTrMapping       *errs.TlMappingError       // Ошибка маппинга структур на транспортном слое
	)

	return errors.As(err, &errInvalidArgument) ||
		errors.As(err, &errBllValidate) ||
		errors.As(err, &errTrMapping)
}

// IsUnauthorized проверяет, вызвана ли ошибка отсутствием, истечением или невалидностью токена/сессии.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 401 (Unauthorized)
//   - В gRPC: статус-код codes.Unauthenticated (16)
func IsUnauthorized(err error) bool {
	var (
		errBllUnauthorized *errs.BllUnauthorizedError // Ошибка аутентификации (протухший JWT, неверный пароль/API-ключ)
	)

	return errors.As(err, &errBllUnauthorized)
}

// IsForbidden проверяет ошибки, связанные с нехваткой прав, ролей или скоупов (RBAC/ABAC) у субъекта запроса.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 403 (Forbidden)
//   - В gRPC: статус-код codes.PermissionDenied (7)
func IsForbidden(err error) bool {
	var (
		errBllForbidden *errs.BllForbiddenError // Личность опознана, но у текущего пользователя нет прав на этот ресурс/действие
	)

	return errors.As(err, &errBllForbidden)
}

// IsNotFound проверяет, вызвана ли ошибка физическим или логическим отсутствием запрашиваемого ресурса в системе.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 404 (Not Found)
//   - В gRPC: статус-код codes.NotFound (5)
func IsNotFound(err error) bool {
	var (
		errBllNotFound *errs.BllNotFoundError // Логическое отсутствие объекта на уровне бизнес-сценария (указан неверный UUID)
		errDalNotFound *errs.DalNotFoundError // Физическое отсутствие строки в базе данных (sql.ErrNoRows на уровне Helper)
	)

	return errors.As(err, &errBllNotFound) ||
		errors.As(err, &errDalNotFound)
}

// IsConflict выявляет ошибки состояния системы, когда операция нарушает консистентность или уникальность данных.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 409 (Conflict)
//   - В gRPC: статус-код codes.AlreadyExists (6) или codes.Aborted (10)
func IsConflict(err error) bool {
	var (
		errBllUnique        *errs.BllUniqueError        // Конфликт уникальности, пойманный на уровне бизнес-логики домена
		errDalAlreadyExists *errs.DalAlreadyExistsError // Нарушение уникального индекса (Unique Violation) на уровне СУБД (PostgreSQL)
	)

	return errors.As(err, &errBllUnique) ||
		errors.As(err, &errDalAlreadyExists)
}

// IsGone проверяет, удален ли запрашиваемый ресурс из системы окончательно и безвозвратно.
// Используется сквозными интерцепторами для маппинга:
//   - В HTTP: статус-код 410 (Gone)
//   - В gRPC: статус-код codes.NotFound (5) или codes.FailedPrecondition (9)
func IsGone(err error) bool {
	var (
		errDalSoftDeleted *errs.DalSoftDeletedError // Ресурс помечен как удаленный (Soft Delete) в DAL-слое базы данных
	)

	return errors.As(err, &errDalSoftDeleted)
}
