package v2

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func dummyEntityFactory() *dummyModel {
	return &dummyModel{}
}

func createValidOpts(t *testing.T) (*mocks.MockExecutor, *mocks.MockErrorDecipher, *Statements[*dummyModel, string]) {
	mockExec := mocks.NewMockExecutor(t)
	mockDecipher := mocks.NewMockErrorDecipher(t)
	stmts := NewStatements[*dummyModel, string]().WithStatement(StatementFind, &Statement[*dummyModel, string]{})
	return mockExec, mockDecipher, stmts
}

func TestNewCRUD_Validation(t *testing.T) {
	t.Run("options validation failure", func(t *testing.T) {
		repo, err := NewCRUD[*dummyModel, string]()
		assert.Error(t, err)
		assert.Nil(t, repo)
		assert.Contains(t, err.Error(), "options validations failed")
	})

	t.Run("success initialization", func(t *testing.T) {
		mockExec, mockDecipher, stmts := createValidOpts(t)

		repo, err := NewCRUD[*dummyModel, string](
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithHooks[*dummyModel, string](WithHookNewEntityFactory[*dummyModel, string](dummyEntityFactory)),
			func(co *CRUDOptions[*dummyModel, string]) {
				co.Statements = stmts
			},
		)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})
}

func TestCRUD_Find(t *testing.T) {
	ctx := context.Background()

	t.Run("validation hook failure", func(t *testing.T) {
		mockExec, mockDecipher, stmts := createValidOpts(t)
		repo, _ := NewCRUD[*dummyModel, string](
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithHooks[*dummyModel, string](
				WithHookNewEntityFactory[*dummyModel, string](dummyEntityFactory),
				WithValidateFind[*dummyModel, string](func(ctx context.Context, id string) error {
					return errors.New("invalid id format")
				}),
			),
			func(co *CRUDOptions[*dummyModel, string]) { co.Statements = stmts },
		)

		_, err := repo.Find(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "find validate hook failed")
	})

	t.Run("success execution", func(t *testing.T) {
		mockExec, mockDecipher, stmts := createValidOpts(t)
		mockQuerier := mocks.NewMockQuerier(t)
		mockExec.On("GetQuerier", ctx).Return(mockQuerier)
		mockQuerier.On("QueryRowContext", ctx, mock.Anything, []any{"123"}).Return(nil)

		repo, _ := NewCRUD[*dummyModel, string](
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithHooks[*dummyModel, string](WithHookNewEntityFactory[*dummyModel, string](dummyEntityFactory)),
			func(co *CRUDOptions[*dummyModel, string]) { co.Statements = stmts },
		)

		repo.helper.hooks.DispatchMapper = func(ctx context.Context, name string, row Scannable, params ...any) (*dummyModel, error) {
			return &dummyModel{ID: "123"}, nil
		}

		res, err := repo.Find(ctx, "123")
		assert.NoError(t, err)
		assert.Equal(t, "123", res.ID)
	})
}

func TestCRUD_List(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid limit", func(t *testing.T) {
		mockExec, mockDecipher, stmts := createValidOpts(t)
		repo, _ := NewCRUD[*dummyModel, string](
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithHooks[*dummyModel, string](WithHookNewEntityFactory[*dummyModel, string](dummyEntityFactory)),
			func(co *CRUDOptions[*dummyModel, string]) { co.Statements = stmts },
		)

		_, err := repo.List(ctx, 0, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be greater 0")
	})
}

func TestCRUD_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("missing creator hook", func(t *testing.T) {
		mockExec, mockDecipher, stmts := createValidOpts(t)
		repo, _ := NewCRUD[*dummyModel, string](
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithHooks[*dummyModel, string](WithHookNewEntityFactory[*dummyModel, string](dummyEntityFactory)),
			func(co *CRUDOptions[*dummyModel, string]) { co.Statements = stmts },
		)

		_, err := repo.Create(ctx, &dummyModel{ID: "1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "creator hook not applied")
	})
}

func TestCRUD_MapperDispatcher(t *testing.T) {
	ctx := context.Background()

	t.Run("mapper path success", func(t *testing.T) {
		mockExec, mockDecipher, _ := createValidOpts(t)
		expectedModel := &dummyModel{ID: "mapped"}

		stmts := NewStatements[*dummyModel, string]().WithStatement("CustomStmt", &Statement[*dummyModel, string]{
			SQL: "SELECT 1",
			Mapper: func(ctx context.Context, scannable Scannable, params ...any) (*dummyModel, error) {
				return expectedModel, nil
			},
		})

		repo, _ := NewCRUD[*dummyModel, string](
			WithExecutor[*dummyModel, string](mockExec),
			WithErrorDecipher[*dummyModel, string](mockDecipher),
			WithEntityInfo[*dummyModel, string]("users", "User"),
			WithHooks[*dummyModel, string](WithHookNewEntityFactory[*dummyModel, string](dummyEntityFactory)),
			func(co *CRUDOptions[*dummyModel, string]) { co.Statements = stmts },
		)

		res, err := repo.mapperDispatcher(ctx, "CustomStmt", nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedModel, res)
	})
}
