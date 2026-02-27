package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	mocks2 "github.com/ElfAstAhe/go-service-template/internal/domain/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestSaveUseCase_CreateAndChange(t *testing.T) {
	// prepare
	inputCreate := domain.NewTest("", "test1", "Test Entity 1", "", time.Now(), time.Now())
	expectedCreate := domain.NewTest("1", "test1", "Test Entity 1", "", time.Now(), time.Now())
	inputChange := domain.NewTest("2", "test2", "Test Entity 2", "", time.Now(), time.Now())
	expectedChange := domain.NewTest("2", "test2", "Test Entity 2", "", time.Now(), time.Now())
	ctx := context.Background()

	tests := []struct {
		name         string
		input        *domain.Test
		prepareMocks func(mRepo *mocks2.MockTestRepository, mTM *mocks.MockTransactionManager)
		expectedRes  *domain.Test
		expectedErr  string // Подстрока ошибки для проверки
	}{
		{
			name:  "Success: entity created",
			input: inputCreate,
			prepareMocks: func(mRepo *mocks2.MockTestRepository, mTM *mocks.MockTransactionManager) {
				// Эмулируем успешную транзакцию
				mTM.On("WithinTransaction", mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						// Вызываем callback-функцию, которую передали в TransactionManager
						fn := args.Get(2).(func(context.Context) error)
						_ = fn(ctx)
					})
				mRepo.On("Create", mock.Anything, inputCreate).Return(expectedCreate, nil)
			},
			expectedRes: expectedCreate,
			expectedErr: "",
		},
		{
			name:  "Success: entity changed",
			input: inputChange,
			prepareMocks: func(mRepo *mocks2.MockTestRepository, mTM *mocks.MockTransactionManager) {
				// эмулируем успешную транзакцию
				mTM.On("WithinTransaction", mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						fn := args.Get(2).(func(context.Context) error)
						_ = fn(ctx)
					})
				mRepo.On("Change", mock.Anything, inputChange).Return(expectedChange, nil)
			},
			expectedRes: expectedChange,
			expectedErr: "",
		},
		{
			name:  "Error: repository failure create",
			input: inputCreate,
			prepareMocks: func(mRepo *mocks2.MockTestRepository, mTM *mocks.MockTransactionManager) {
				mTM.On("WithinTransaction", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("db error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(2).(func(context.Context) error)
						_ = fn(ctx)
					})
				mRepo.On("Create", mock.Anything, inputCreate).Return(nil, errors.New("sql fail"))
			},
			expectedRes: nil,
			expectedErr: "save test model",
		},
		{
			name:  "Error: repository failure change",
			input: inputChange,
			prepareMocks: func(mRepo *mocks2.MockTestRepository, mTM *mocks.MockTransactionManager) {
				mTM.On("WithinTransaction", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("db error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(2).(func(context.Context) error)
						_ = fn(ctx)
					})
				mRepo.On("Change", mock.Anything, inputChange).Return(nil, errors.New("sql fail"))
			},
			expectedRes: nil,
			expectedErr: "save test model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			mRepo := new(mocks2.MockTestRepository)
			mTM := new(mocks.MockTransactionManager)
			tt.prepareMocks(mRepo, mTM)

			uc := NewTestSaveUseCase(mTM, mRepo)

			// act
			actual, err := uc.Save(ctx, tt.input)

			// assert
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, actual)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, actual)
			}

			mRepo.AssertExpectations(t)
			mTM.AssertExpectations(t)
		})
	}
}
