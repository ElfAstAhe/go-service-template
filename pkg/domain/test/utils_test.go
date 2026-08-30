package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEntitiesToIDList_WithMocks(t *testing.T) {
	t.Run("should extract int IDs", func(t *testing.T) {
		// Создаем моки для int ID
		ent1 := mocks.NewMockEntity[int](t)
		ent2 := mocks.NewMockEntity[int](t)

		// Настраиваем ожидания (Type-safe способ через EXPECT)
		ent1.EXPECT().GetID().Return(1).Once()
		ent2.EXPECT().GetID().Return(2).Once()

		src := []domain.Entity[int]{ent1, ent2}

		res := domain.EntitiesToIDList(src)

		assert.Equal(t, []int{1, 2}, res)
	})

	t.Run("should extract string IDs", func(t *testing.T) {
		// Создаем моки для string ID
		ent := mocks.NewMockEntity[string](t)

		ent.EXPECT().GetID().Return("uuid-123").Once()

		src := []domain.Entity[string]{ent}

		res := domain.EntitiesToIDList(src)

		assert.Equal(t, []string{"uuid-123"}, res)
	})
}

func TestAssignUUIDv7(t *testing.T) {
	t.Run("успешная генерация и присвоение UUIDv7", func(t *testing.T) {
		// 1. Создаем мок сущности с типом string для ID
		mockEntity := mocks.NewMockEntity[string](t)

		// 2. Настраиваем ожидания: метод SetID должен быть вызван
		// с любой строкой (так как UUID каждый раз новый), и ничего не возвращает.
		mockEntity.On("SetID", mock.AnythingOfType("string")).Once()

		// 3. Вызываем тестируемый метод
		err := domain.AssignUUIDv7(mockEntity)

		// 4. Проверяем, что ошибки нет
		assert.NoError(t, err)

		// На всякий случай проверяем, что переданный аргумент в SetID
		// действительно похож на валидный UUID (опционально, для строгой проверки)
		call := mockEntity.Calls[0]
		assignedID := call.Arguments.String(0)
		assert.Len(t, assignedID, 36, "ID должен быть длиной 36 символов (стандартный UUID)")
	})
}

func TestAssignUUIDv4(t *testing.T) {
	t.Run("успешная генерация и присвоение UUIDv4", func(t *testing.T) {
		// 1. Создаем мок сущности с типом string для ID
		mockEntity := mocks.NewMockEntity[string](t)

		// 2. Настраиваем ожидания: метод SetID должен быть вызван
		// с любой строкой (так как UUID каждый раз новый), и ничего не возвращает.
		mockEntity.On("SetID", mock.AnythingOfType("string")).Once()

		// 3. Вызываем тестируемый метод
		err := domain.AssignUUIDv4(mockEntity)

		// 4. Проверяем, что ошибки нет
		assert.NoError(t, err)

		// На всякий случай проверяем, что переданный аргумент в SetID
		// действительно похож на валидный UUID (опционально, для строгой проверки)
		call := mockEntity.Calls[0]
		assignedID := call.Arguments.String(0)
		assert.Len(t, assignedID, 36, "ID должен быть длиной 36 символов (стандартный UUID)")
	})
}
