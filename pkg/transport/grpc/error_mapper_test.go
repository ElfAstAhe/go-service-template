package grpc

import (
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMapToGrpcError осуществляет комплексную верификацию маппинга ошибок фреймворка,
// используя оригинальные специализированные типы пакета errs для 100% покрытия контракта transport.
func TestMapToGrpcError(t *testing.T) {
	// 1. Аллоцируем реальные специализированные типы ошибок из твоего пакета errs
	badRequestErr := errs.NewInvalidArgumentError("id", "field id is malformed")

	// Конструируем структуры ошибок под жесткие типы, прописанные в errors.As пакета transport.
	// Используем те фабричные методы или структуры пакета errs, которые создают именно указанные типы.
	// ПРИМЕЧАНИЕ: Если в твоем пакете errs для этих типов есть явные фабричные функции (NewBllUnauthorizedError и т.д.),
	// используй их. Ниже приведен классический пример инициализации под типы из твоего файла конфигурации:
	unauthorizedErr := errs.NewBllUnauthorizedError("Auth", "token expired", nil)
	forbiddenErr := errs.NewBllForbiddenError("Auth", "access denied", nil)
	notFoundErr := errs.NewBllNotFoundError("DAL", "test", "id-012345", nil)
	conflictErr := errs.NewBllUniqueError("DAL", "test", "id-12345", nil)
	goneErr := errs.NewDalSoftDeletedError("DAL", "id-123455")

	// Описываем структуру табличного тест-кейса с динамической сверкой строк
	tests := []struct {
		name         string     // Понятное описание сценария
		inputErr     error      // Входящая абстрактная ошибка
		expectedCode codes.Code // Ожидаемый сетевой gRPC-код ответа
		expectedMsg  string     // Ожидаемый текст сообщения в gRPC статусе
	}{
		{
			name:         "Успешный пропуск nil-ошибки без генерации сетевого статуса",
			inputErr:     nil,
			expectedCode: codes.OK,
			expectedMsg:  "",
		},
		{
			name:         "Успешная трансляция BadRequest ошибки в codes.InvalidArgument",
			inputErr:     badRequestErr,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  badRequestErr.Error(), // Динамически берем текст со всеми префиксами платформы
		},
		{
			name:         "Успешная трансляция Unauthorized ошибки в codes.Unauthenticated",
			inputErr:     unauthorizedErr,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  unauthorizedErr.Error(),
		},
		{
			name:         "Успешная трансляция Forbidden ошибки в codes.PermissionDenied",
			inputErr:     forbiddenErr,
			expectedCode: codes.PermissionDenied,
			expectedMsg:  forbiddenErr.Error(),
		},
		{
			name:         "Успешная трансляция NotFound ошибки в codes.NotFound",
			inputErr:     notFoundErr,
			expectedCode: codes.NotFound,
			expectedMsg:  notFoundErr.Error(),
		},
		{
			name:         "Успешная трансляция Conflict ошибки в codes.AlreadyExists",
			inputErr:     conflictErr,
			expectedCode: codes.AlreadyExists,
			expectedMsg:  conflictErr.Error(),
		},
		{
			name:         "Успешная трансляция логически удаленного ресурса Gone в codes.NotFound",
			inputErr:     goneErr,
			expectedCode: codes.NotFound,
			expectedMsg:  goneErr.Error(),
		},
		{
			name:         "Предохранитель: маскирование сырой системной ошибки в константный codes.Internal",
			inputErr:     errors.New("fatal: postgres connection pool exhausted"),
			expectedCode: codes.Internal,
			expectedMsg:  "internal server error", // Сырой текст системного сбоя скрывается
		},
	}

	// Запускаем итерационный цикл прогона таблицы
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resErr := MapToGrpcError(tt.inputErr)

			// Если на входе был nil, на выходе обязаны строго получить nil-error
			if tt.inputErr == nil {
				assert.NoError(t, resErr)
				return
			}

			// Извлекаем gRPC-статус из возвращаемой сетевой ошибки с помощью пакета status
			grpcStatus, ok := status.FromError(resErr)
			assert.True(t, ok, "Возвращаемая ошибка должна быть валидным gRPC-статусом")

			// Верифицируем сетевой код ответа и защитное сообщение
			assert.Equal(t, tt.expectedCode, grpcStatus.Code(), "Неверный gRPC-код ответа")
			assert.Equal(t, tt.expectedMsg, grpcStatus.Message(), "Текст ошибки gRPC-статуса не совпадает")
		})
	}
}
