package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/domain/mocks"
	"github.com/stretchr/testify/assert"
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
