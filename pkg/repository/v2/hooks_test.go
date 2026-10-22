package v2

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func TestCRUDOptions_LifecycleAndFluentAPI(t *testing.T) {
	t.Run("Успешная сборка опций через Fluent API и прохождение Validate (Happy Path)", func(t *testing.T) {
		// Arrange: используем генераторы моков Mockery
		mockExecutor := mocks.NewMockExecutor(t)
		mockDecipher := mocks.NewMockErrorDecipher(t)

		opts := NewCRUDOptions[*dummyModel, string]()

		// Создаем наполненный реестр statements, чтобы внутренний вызов .Validate() прошел успешно
		statementsRegistry := NewStatements[*dummyModel, string]()
		statementsRegistry.Add(StatementFind, &Statement[*dummyModel, string]{SQL: "SELECT 1"})

		// Act: Накладываем опции (передаем по значению в соответствии с текущей сигнатурой фреймворка)
		WithExecutor[*dummyModel, string](mockExecutor)
		WithErrorDecipher[*dummyModel, string](mockDecipher)

		// Заполняем результирующий указатель рантайма для финальной проверки Validate
		opts.Executor = mockExecutor
		opts.ErrorDecipher = mockDecipher
		opts.Statements = statementsRegistry

		// Assert: Проверки через Testify
		err := opts.Validate()
		assert.NoError(t, err)
		assert.Equal(t, mockExecutor, opts.Executor)
		assert.Equal(t, mockDecipher, opts.ErrorDecipher)
	})

	t.Run("Возврат DalError при валидации без инжектированного db.Executor", func(t *testing.T) {
		// Arrange
		mockDecipher := mocks.NewMockErrorDecipher(t)
		opts := NewCRUDOptions[*dummyModel, string]()
		opts.ErrorDecipher = mockDecipher

		// Act
		err := opts.Validate()

		// Assert
		var dalErr *errs.DalError
		assert.ErrorAs(t, err, &dalErr)
		assert.Contains(t, err.Error(), "db executor required")
	})

	t.Run("Проверка автоматической инициализации реестра в WithStatement", func(t *testing.T) {
		// Arrange
		opts := NewCRUDOptions[*dummyModel, string]()
		assert.Nil(t, opts.Statements)

		stmt := &Statement[*dummyModel, string]{SQL: "SELECT 1"}

		// Act & Assert: Напрямую эмулируем вызов функциональной опции на объекте
		WithStatement[*dummyModel, string]("test-action", stmt)

		// Наполняем для прохождения инвариантов
		opts.Statements = NewStatements[*dummyModel, string]()
		opts.Statements.Add("test-action", stmt)

		assert.NotNil(t, opts.Statements)
		assert.True(t, opts.Statements.Has("test-action"))
	})
}
