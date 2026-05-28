package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/container/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetInstance(t *testing.T) {
	// Создаем мок контейнера
	mockCtn := mocks.NewMockContainer(t)
	mockCtn.On("GetName").Return("test-container").Maybe()

	t.Run("Success_String", func(t *testing.T) {
		name := "string-instance"
		expected := "hello-world"

		mockCtn.On("GetInstance", name).Return(expected, nil).Once()

		res, err := container.GetInstance[string](mockCtn, name)

		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("Success_Interface", func(t *testing.T) {
		name := "complex-instance"
		// Предположим, мы ждем какой-то интерфейс или структуру
		type myStruct struct{ ID int }
		expected := &myStruct{ID: 42}

		mockCtn.On("GetInstance", name).Return(expected, nil).Once()

		res, err := container.GetInstance[*myStruct](mockCtn, name)

		assert.NoError(t, err)
		assert.Equal(t, expected.ID, res.ID)
	})

	t.Run("Error_TypeMismatch", func(t *testing.T) {
		name := "wrong-type"
		instance := 123 // В контейнере число

		mockCtn.On("GetInstance", name).Return(instance, nil).Once()

		// А мы пытаемся достать его как строку
		res, err := container.GetInstance[string](mockCtn, name)

		assert.Error(t, err)
		assert.Empty(t, res)
		assert.Contains(t, err.Error(), "instance type")
		assert.Contains(t, err.Error(), "mismatch")
	})

	t.Run("Handle_Nil_Instance", func(t *testing.T) {
		name := "nil-key"
		// Имитируем, что GetInstance вернул nil (через интерфейс)
		mockCtn.On("GetInstance", name).Return(nil, nil).Once()

		res, err := container.GetInstance[*testing.T](mockCtn, name)

		assert.NoError(t, err)
		assert.Nil(t, res, "Should return typed nil without error if instance is nil")
	})

	t.Run("Error_ContainerNil", func(t *testing.T) {
		res, err := container.GetInstance[string](nil, "any")

		assert.Error(t, err)
		assert.Empty(t, res)
		assert.Contains(t, err.Error(), "container nil")
	})

	t.Run("Error_EmptyName", func(t *testing.T) {
		res, err := container.GetInstance[string](mockCtn, "")

		assert.Error(t, err)
		assert.Empty(t, res)
		assert.Contains(t, err.Error(), "instance name is empty")
	})

	t.Run("Error_FromContainer", func(t *testing.T) {
		name := "error-key"
		// Имитируем, что сам контейнер вернул ошибку (например, Not Found)
		mockCtn.On("GetInstance", name).Return(nil, assert.AnError).Once()

		res, err := container.GetInstance[string](mockCtn, name)

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		assert.Empty(t, res)
	})
}
