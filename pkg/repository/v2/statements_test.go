package v2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Фиктивная функция маппера для удовлетворения сигнатуры EntityMapperFunc
func dummyMapper(ctx context.Context, scannable Scannable, params ...any) (*dummyModel, error) {
	return &dummyModel{}, nil
}

func TestNewStatement(t *testing.T) {
	sqlStr := "SELECT * FROM users"
	stmt := NewStatement[*dummyModel, string](sqlStr, dummyMapper)

	assert.NotNil(t, stmt)
	assert.Equal(t, sqlStr, stmt.SQL)
	assert.NotNil(t, stmt.Mapper)
}

func TestStatements_CRUDAndGetters(t *testing.T) {
	// Инициализация пустого реестра
	stmts := NewStatements[*dummyModel, string]()
	assert.NotNil(t, stmts)
	assert.False(t, stmts.Has("GetUser"))

	// Проверка Validate на пустом реестре
	err := stmts.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "statements required")

	// Тестирование Add и Has
	stmt1 := NewStatement[*dummyModel, string]("SELECT * FROM users WHERE id = ?", dummyMapper)
	stmts.Add("GetUser", stmt1)
	assert.True(t, stmts.Has("GetUser"))

	// Проверка Validate на непустом реестре
	err = stmts.Validate()
	assert.NoError(t, err)

	// Тестирование Get и экспресс-геттеров
	gotStmt, err := stmts.Get("GetUser")
	assert.NoError(t, err)
	assert.Equal(t, stmt1, gotStmt)

	gotSQL, err := stmts.GetSQL("GetUser")
	assert.NoError(t, err)
	assert.Equal(t, stmt1.SQL, gotSQL)

	gotMapper, err := stmts.GetMapper("GetUser")
	assert.NoError(t, err)
	assert.NotNil(t, gotMapper)

	// Тестирование ошибок при запросе несуществующего стейтмента
	_, err = stmts.Get("Unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "statement mapper Unknown not found")

	_, err = stmts.GetSQL("Unknown")
	assert.Error(t, err)

	_, err = stmts.GetMapper("Unknown")
	assert.Error(t, err)

	// Тестирование Fluent API: WithStatement
	stmt2 := NewStatement[*dummyModel, string]("DELETE FROM users", dummyMapper)
	returnedStmts := stmts.WithStatement("DeleteUser", stmt2)
	assert.Equal(t, stmts, returnedStmts)
	assert.True(t, stmts.Has("DeleteUser"))

	// Тестирование Remove
	stmts.Remove("GetUser")
	assert.False(t, stmts.Has("GetUser"))
}
