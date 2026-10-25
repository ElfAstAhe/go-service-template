package v2

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewCRUDOptions(t *testing.T) {
	opts := NewCRUDOptions[*dummyModel, string]()

	assert.NotNil(t, opts)
	assert.Nil(t, opts.Executor)
	assert.Nil(t, opts.ErrorDecipher)
	assert.Nil(t, opts.Statements)
	assert.Empty(t, opts.Hooks)
	assert.Empty(t, opts.TableName)
	assert.Empty(t, opts.EntityName)
}

func TestCRUDOptions_FluentAPI(t *testing.T) {
	// Arrange
	mockExec := mocks.NewMockExecutor(t)
	mockDecipher := mocks.NewMockErrorDecipher(t)

	dummyStmt := &Statement[*dummyModel, string]{}
	var dummyHook Hook[*dummyModel, string]

	opts := NewCRUDOptions[*dummyModel, string]()

	// Act & Assert
	WithExecutor[*dummyModel, string](mockExec)(opts)
	assert.Equal(t, mockExec, opts.Executor)

	WithErrorDecipher[*dummyModel, string](mockDecipher)(opts)
	assert.Equal(t, mockDecipher, opts.ErrorDecipher)

	WithEntityInfo[*dummyModel, string]("users", "User")(opts)
	assert.Equal(t, "users", opts.TableName)
	assert.Equal(t, "User", opts.EntityName)

	WithHooks[*dummyModel, string](dummyHook)(opts)
	assert.Len(t, opts.Hooks, 1)

	// Проверяем ленивую инициализацию реестра внутри WithStatement
	assert.Nil(t, opts.Statements)
	WithStatement[*dummyModel, string]("FindByID", dummyStmt)(opts)
	assert.NotNil(t, opts.Statements)
}

func TestCRUDOptions_Validate(t *testing.T) {
	mockExec := mocks.NewMockExecutor(t)
	mockDecipher := mocks.NewMockErrorDecipher(t)
	validStatements := NewStatements[*dummyModel, string]()
	validStatements.Add("test sql", nil)

	t.Run("success path", func(t *testing.T) {
		opts := &CRUDOptions[*dummyModel, string]{
			Executor:      mockExec,
			ErrorDecipher: mockDecipher,
			Statements:    validStatements,
			TableName:     "products",
			EntityName:    "Product",
		}

		err := opts.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing executor", func(t *testing.T) {
		opts := &CRUDOptions[*dummyModel, string]{
			Executor:      nil,
			ErrorDecipher: mockDecipher,
			Statements:    validStatements,
			TableName:     "products",
			EntityName:    "Product",
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db executor required")
	})

	t.Run("missing error decipher", func(t *testing.T) {
		opts := &CRUDOptions[*dummyModel, string]{
			Executor:      mockExec,
			ErrorDecipher: nil,
			Statements:    validStatements,
			TableName:     "products",
			EntityName:    "Product",
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error decipher required")
	})

	t.Run("empty table name", func(t *testing.T) {
		opts := &CRUDOptions[*dummyModel, string]{
			Executor:      mockExec,
			ErrorDecipher: mockDecipher,
			Statements:    validStatements,
			TableName:     "   ", // Проверка strings.TrimSpace
			EntityName:    "Product",
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table name required")
	})

	t.Run("empty entity name", func(t *testing.T) {
		opts := &CRUDOptions[*dummyModel, string]{
			Executor:      mockExec,
			ErrorDecipher: mockDecipher,
			Statements:    validStatements,
			TableName:     "products",
			EntityName:    "",
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entity name required")
	})
}
