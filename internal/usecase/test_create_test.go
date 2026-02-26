package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	mocks2 "github.com/ElfAstAhe/go-service-template/internal/domain/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestCreateUseCase_Create(t *testing.T) {
	// Инициализация общих данных
	testInput := &domain.Test{ID: "1", Name: "Test Entity"}
	ctx := context.Background()

	// Описываем таблицу тестов
	tests := []struct {
		name         string
		input        *domain.Test
		prepareMocks func(mRepo *mocks2.MockTestRepository, mTM *mocks.MockTransactionManager)
		expectedRes  *domain.Test
		expectedErr  string // Подстрока ошибки для проверки
	}{
		{
			name:  "Success: entity created",
			input: testInput,
			prepareMocks: func(mRepo *mocks2.MockTestRepository, mTM *mocks.MockTransactionManager) {
				// Эмулируем успешную транзакцию
				mTM.On("WithinTransaction", mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						// Вызываем callback-функцию, которую передали в TransactionManager
						fn := args.Get(2).(func(context.Context) error)
						_ = fn(ctx)
					})
				mRepo.On("Create", mock.Anything, testInput).Return(testInput, nil)
			},
			expectedRes: testInput,
			expectedErr: "",
		},
		{
			name:  "Error: repository failure",
			input: testInput,
			prepareMocks: func(mRepo *mocks2.MockTestRepository, mTM *mocks.MockTransactionManager) {
				mTM.On("WithinTransaction", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("db error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(2).(func(context.Context) error)
						_ = fn(ctx)
					})
				mRepo.On("Create", mock.Anything, testInput).Return(nil, errors.New("sql fail"))
			},
			expectedRes: nil,
			expectedErr: "error create test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Setup
			mRepo := new(mocks2.MockTestRepository)
			mTM := new(mocks.MockTransactionManager)
			tt.prepareMocks(mRepo, mTM)

			uc := NewTestCreateUseCase(mTM, mRepo)

			// 2. Execute
			res, err := uc.Create(ctx, tt.input)

			// 3. Assert
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}

			mRepo.AssertExpectations(t)
			mTM.AssertExpectations(t)
		})
	}
}
