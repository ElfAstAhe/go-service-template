package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUnitOfWork_Execute_WithMocks(t *testing.T) {
	t.Run("Успешное выполнение бизнес-логики внутри транзакции (Happy Path)", func(t *testing.T) {
		// Arrange
		mockTM := mocks.NewMockTransactionManager(t)
		ctx := context.Background()
		uow := NewUnitOfWork(mockTM, nil)

		// Настраиваем существующий мок через Mockery контракт
		mockTM.On("WithinTransaction", ctx, uow.defaultTMO, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
				// Выполняем саму функцию бизнес-логики, переданную аргументом
				fn := args.Get(2).(func(context.Context) error)
				_ = fn(ctx)
			}).
			Once()

		businessLogicCalled := false
		dummyLogic := func(c context.Context) error {
			businessLogicCalled = true
			return nil
		}

		// Act
		err := uow.Execute(ctx, dummyLogic)

		// Assert
		assert.NoError(t, err)
		assert.True(t, businessLogicCalled, "Бизнес-логика должна быть успешно вызвана")
	})

	t.Run("Транзакция возвращает ошибку, если бизнес-логика упала (Rollback сценарий)", func(t *testing.T) {
		// Arrange
		mockTM := mocks.NewMockTransactionManager(t)
		ctx := context.Background()
		uow := NewUnitOfWork(mockTM, nil)
		expectedErr := errors.New("ошибка бизнес-логики")

		mockTM.On("WithinTransaction", ctx, uow.defaultTMO, mock.AnythingOfType("func(context.Context) error")).
			Return(expectedErr).
			Once()

		dummyLogic := func(c context.Context) error {
			return expectedErr
		}

		// Act
		err := uow.Execute(ctx, dummyLogic)

		// Assert
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("Использование кастомных опций транзакции при их наличии", func(t *testing.T) {
		// Arrange
		mockTM := mocks.NewMockTransactionManager(t)
		ctx := context.Background()
		customOpts := &db.TransactionOptions{
			ReadOnly: true,
		}

		uow := NewUnitOfWork(mockTM, customOpts)

		mockTM.On("WithinTransaction", ctx, customOpts, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Once()

		// Act
		err := uow.Execute(ctx, func(c context.Context) error { return nil })

		// Assert
		assert.NoError(t, err)
	})
}
