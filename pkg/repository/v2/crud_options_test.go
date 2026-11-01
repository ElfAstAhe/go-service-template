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
	assert.Empty(t, opts.Hooks)
	assert.Empty(t, opts.TableName)
	assert.Empty(t, opts.EntityName)
	assert.Empty(t, opts.DriverName)
}

func TestCRUDOptions_FluentAPI(t *testing.T) {
	// Arrange
	mockExec := mocks.NewMockExecutor(t)
	mockDecipher := mocks.NewMockErrorDecipher(t)

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
}

func TestCRUDOptions_Validate(t *testing.T) {
	mockExec := mocks.NewMockExecutor(t)
	mockDecipher := mocks.NewMockErrorDecipher(t)

	t.Run("success path", func(t *testing.T) {
		optList := []CRUDOption[*dummyModel, string]{
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithDriverName[*dummyModel, string]("dummy"),
			WithStatementsBuilder[*dummyModel, string]("dummy", func() *Statements[*dummyModel, string] {
				return NewStatements[*dummyModel, string]()
			}),
		}
		opts := NewCRUDOptions[*dummyModel, string]()
		for _, opt := range optList {
			opt(opts)
		}

		err := opts.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing executor", func(t *testing.T) {
		optList := []CRUDOption[*dummyModel, string]{
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithDriverName[*dummyModel, string]("dummy"),
			WithStatementsBuilder[*dummyModel, string]("dummy", func() *Statements[*dummyModel, string] {
				return NewStatements[*dummyModel, string]()
			}),
		}
		opts := NewCRUDOptions[*dummyModel, string]()
		for _, opt := range optList {
			opt(opts)
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db executor required")
	})

	t.Run("missing error decipher", func(t *testing.T) {
		optList := []CRUDOption[*dummyModel, string]{
			WithExecutor[*dummyModel, string](mockExec),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithDriverName[*dummyModel, string]("dummy"),
			WithStatementsBuilder[*dummyModel, string]("dummy", func() *Statements[*dummyModel, string] {
				return NewStatements[*dummyModel, string]()
			}),
		}
		opts := NewCRUDOptions[*dummyModel, string]()
		for _, opt := range optList {
			opt(opts)
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error decipher required")
	})

	t.Run("empty table name", func(t *testing.T) {
		optList := []CRUDOption[*dummyModel, string]{
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("   ", "User"),
			WithDriverName[*dummyModel, string]("dummy"),
			WithStatementsBuilder[*dummyModel, string]("dummy", func() *Statements[*dummyModel, string] {
				return NewStatements[*dummyModel, string]()
			}),
		}
		opts := NewCRUDOptions[*dummyModel, string]()
		for _, opt := range optList {
			opt(opts)
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table name required")
	})

	t.Run("empty entity name", func(t *testing.T) {
		optList := []CRUDOption[*dummyModel, string]{
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "   "),
			WithDriverName[*dummyModel, string]("dummy"),
			WithStatementsBuilder[*dummyModel, string]("dummy", func() *Statements[*dummyModel, string] {
				return NewStatements[*dummyModel, string]()
			}),
		}
		opts := NewCRUDOptions[*dummyModel, string]()
		for _, opt := range optList {
			opt(opts)
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entity name required")
	})

	t.Run("statement builders required", func(t *testing.T) {
		optList := []CRUDOption[*dummyModel, string]{
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithDriverName[*dummyModel, string]("dummy"),
		}
		opts := NewCRUDOptions[*dummyModel, string]()
		for _, opt := range optList {
			opt(opts)
		}

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "driver statements builders required")
	})
}
