package v2

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Ручная заглушка для sql.Result, чтобы не мокать std-библиотеку
type mockSqlResult struct {
	rowsAffected int64
	lastInsertId int64
	err          error
}

func (m mockSqlResult) LastInsertId() (int64, error) { return m.lastInsertId, m.err }
func (m mockSqlResult) RowsAffected() (int64, error) { return m.rowsAffected, m.err }

func TestHelper_Get(t *testing.T) {
	ctx := context.Background()
	stmtName := "FindByID"
	sqlQuery := "SELECT id FROM users WHERE id = ?"

	t.Run("success path with after find hook", func(t *testing.T) {
		mockExec := mocks.NewMockExecutor(t)
		mockQuerier := mocks.NewMockQuerier(t)
		mockDecipher := mocks.NewMockErrorDecipher(t)

		mockExec.On("GetQuerier", ctx).Return(mockQuerier)
		// Настраиваем QueryRowContext на возврат чего-либо (DispatchMapper перехватит вызов)
		mockQuerier.On("QueryRowContext", ctx, sqlQuery, []any{"123"}).Return(&sql.Row{})

		expectedEntity := dummyModel{ID: "123"}
		hookCalled := false

		hHooks := &Hooks[*dummyModel, string]{
			DispatchMapper: func(ctx context.Context, name string, row Scannable, params ...any) (*dummyModel, error) {
				return &expectedEntity, nil
			},
			AfterFind: func(ctx context.Context, ent *dummyModel, params ...any) (*dummyModel, error) {
				hookCalled = true
				ent.ID = "123-modified"
				return ent, nil
			},
		}

		hStatements := NewStatements[*dummyModel, string]().WithStatement(stmtName, &Statement[*dummyModel, string]{SQL: sqlQuery})

		h := NewHelper(mockExec, mockDecipher, hHooks, hStatements, "users", "User")

		res, err := h.Get(ctx, stmtName, "123")
		assert.NoError(t, err)
		assert.Equal(t, "123-modified", res.ID)
		assert.True(t, hookCalled)
	})

	t.Run("statement not found", func(t *testing.T) {
		mockExec := mocks.NewMockExecutor(t)
		mockQuerier := mocks.NewMockQuerier(t)
		mockExec.On("GetQuerier", ctx).Return(mockQuerier)

		hStatements := NewStatements[*dummyModel, string]() // пустой реестр

		h := NewHelper(mockExec, nil, &Hooks[*dummyModel, string]{}, hStatements, "users", "User")

		_, err := h.Get(ctx, "UnknownStmt", "123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sql statement UnknownStmt not found")
	})
}

func TestHelper_Create(t *testing.T) {
	ctx := context.Background()
	inputEntity := dummyModel{ID: "new-id"}
	stmtName := "InsertUser"

	t.Run("success path with before create hook", func(t *testing.T) {
		mockExec := mocks.NewMockExecutor(t)
		mockQuerier := mocks.NewMockQuerier(t)
		mockExec.On("GetQuerier", ctx).Return(mockQuerier)

		beforeCreateCalled := false
		creatorCalled := false

		hHooks := &Hooks[*dummyModel, string]{
			BeforeCreate: func(ctx context.Context, ent *dummyModel, params ...any) (*dummyModel, error) {
				beforeCreateCalled = true
				return ent, nil
			},
			Creator: func(ctx context.Context, q db.Querier, ent *dummyModel, params ...any) (Scannable, error) {
				creatorCalled = true
				return &sql.Row{}, nil
			},
			DispatchMapper: func(ctx context.Context, name string, row Scannable, params ...any) (*dummyModel, error) {
				return &inputEntity, nil
			},
		}

		hStatements := NewStatements[*dummyModel, string]() // Для Create GetSQL не вызывается напрямую, идет работа через Creator

		h := NewHelper(mockExec, nil, hHooks, hStatements, "users", "User")

		res, err := h.Create(ctx, stmtName, &inputEntity)

		assert.NoError(t, err)
		assert.Equal(t, inputEntity, *res)
		assert.True(t, beforeCreateCalled)
		assert.True(t, creatorCalled)
	})
}

func TestHelper_Change(t *testing.T) {
	ctx := context.Background()
	inputEntity := dummyModel{ID: "existing-id"}
	stmtName := "UpdateUser"

	t.Run("success path", func(t *testing.T) {
		mockExec := mocks.NewMockExecutor(t)
		mockQuerier := mocks.NewMockQuerier(t)
		mockExec.On("GetQuerier", ctx).Return(mockQuerier)

		hHooks := &Hooks[*dummyModel, string]{
			BeforeChange: func(ctx context.Context, ent *dummyModel, params ...any) (*dummyModel, error) {
				return ent, nil
			},
			Changer: func(ctx context.Context, q db.Querier, ent *dummyModel, params ...any) (Scannable, error) {
				return &sql.Row{}, nil
			},
			DispatchMapper: func(ctx context.Context, name string, scannable Scannable, params ...any) (*dummyModel, error) {
				return &inputEntity, nil
			},
		}

		h := NewHelper(mockExec, nil, hHooks, NewStatements[*dummyModel, string](), "users", "User")

		res, err := h.Change(ctx, stmtName, &inputEntity)
		assert.NoError(t, err)
		assert.Equal(t, inputEntity, *res)
	})
}

func TestHelper_Delete(t *testing.T) {
	ctx := context.Background()
	stmtName := "DeleteUser"
	sqlQuery := "DELETE FROM users WHERE id = ?"

	t.Run("success delete", func(t *testing.T) {
		mockExec := mocks.NewMockExecutor(t)
		mockQuerier := mocks.NewMockQuerier(t)
		mockExec.On("GetQuerier", ctx).Return(mockQuerier)

		fakeResult := mockSqlResult{rowsAffected: 1}
		mockQuerier.On("ExecContext", ctx, sqlQuery, []any{"123"}).Return(fakeResult, nil)

		hStatements := NewStatements[*dummyModel, string]().WithStatement(stmtName, &Statement[*dummyModel, string]{SQL: sqlQuery})
		h := NewHelper(mockExec, nil, &Hooks[*dummyModel, string]{}, hStatements, "users", "User")

		err := h.Delete(ctx, stmtName, "123")
		assert.NoError(t, err)
	})

	t.Run("delete rows affected is 0 (not found)", func(t *testing.T) {
		mockExec := mocks.NewMockExecutor(t)
		mockQuerier := mocks.NewMockQuerier(t)
		mockExec.On("GetQuerier", ctx).Return(mockQuerier)

		fakeResult := mockSqlResult{rowsAffected: 0}
		mockQuerier.On("ExecContext", ctx, sqlQuery, []any{"999"}).Return(fakeResult, nil)

		hStatements := NewStatements[*dummyModel, string]().WithStatement(stmtName, &Statement[*dummyModel, string]{SQL: sqlQuery})
		h := NewHelper(mockExec, nil, &Hooks[*dummyModel, string]{}, hStatements, "users", "User")

		err := h.Delete(ctx, stmtName, "999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestHelper_DeleteNoCheck(t *testing.T) {
	ctx := context.Background()
	stmtName := "DeleteUserNoCheck"
	sqlQuery := "DELETE FROM users"

	t.Run("blind delete success", func(t *testing.T) {
		mockExec := mocks.NewMockExecutor(t)
		mockQuerier := mocks.NewMockQuerier(t)
		mockExec.On("GetQuerier", ctx).Return(mockQuerier)

		mockQuerier.On("ExecContext", ctx, sqlQuery, mock.Anything).Return(mockSqlResult{rowsAffected: 0}, nil)

		hStatements := NewStatements[*dummyModel, string]().WithStatement(stmtName, &Statement[*dummyModel, string]{SQL: sqlQuery})
		h := NewHelper(mockExec, nil, &Hooks[*dummyModel, string]{}, hStatements, "users", "User")

		err := h.DeleteNoCheck(ctx, stmtName, "123")
		assert.NoError(t, err)
	})
}
