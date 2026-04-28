package test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxManager_WithinTransaction(t *testing.T) {
	// Инициализируем sqlmock
	sqlDB, mockSql, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Создаем мок нашего интерфейса DB через mockery
	mockDB := mocks.NewMockDB(t)
	// Настраиваем его так, чтобы он всегда возвращал наш sqlmock инстанс
	mockDB.On("GetDB").Return(sqlDB)

	tm := db.NewTxManager(mockDB)

	t.Run("Success_Commit", func(t *testing.T) {
		mockSql.ExpectBegin()
		mockSql.ExpectCommit()

		err := tm.WithinTransaction(context.Background(), nil, func(ctx context.Context) error {
			// Проверяем, что транзакция попала в контекст
			tx := db.GetTx(ctx)
			assert.NotNil(t, tx)
			return nil
		})

		assert.NoError(t, err)
		assert.NoError(t, mockSql.ExpectationsWereMet())
	})

	t.Run("Error_Rollback", func(t *testing.T) {
		mockSql.ExpectBegin()
		mockSql.ExpectRollback()

		expectedErr := errors.New("business_logic_error")
		err := tm.WithinTransaction(context.Background(), nil, func(ctx context.Context) error {
			return expectedErr
		})

		assert.ErrorIs(t, err, expectedErr)
		assert.NoError(t, mockSql.ExpectationsWereMet())
	})

	t.Run("Panic_Recovery_Rollback", func(t *testing.T) {
		mockSql.ExpectBegin()
		mockSql.ExpectRollback()

		err := tm.WithinTransaction(context.Background(), nil, func(ctx context.Context) error {
			panic("something exploded")
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic recovery")
		assert.NoError(t, mockSql.ExpectationsWereMet())
	})

	t.Run("Nested_Calls_Reuse_Transaction", func(t *testing.T) {
		// Ожидаем только один BEGIN/COMMIT, так как второй вызов вложенный
		mockSql.ExpectBegin()
		mockSql.ExpectCommit()

		err := tm.WithinTransaction(context.Background(), nil, func(ctx context.Context) error {
			return tm.WithinTransaction(ctx, nil, func(ctx context.Context) error {
				return nil
			})
		})

		assert.NoError(t, err)
		assert.NoError(t, mockSql.ExpectationsWereMet())
	})
}
