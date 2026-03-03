package usecase

import (
	"context"
	"errors"
	"testing"

	mocks2 "github.com/ElfAstAhe/go-service-template/internal/domain/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestGetUseCase_Delete(t *testing.T) {
	// prepare
	inputSuccess := "1"
	inputFail := "2"
	ctx := context.Background()

	tests := []struct {
		name         string
		input        string
		prepareMocks func(mTM *mocks.MockTransactionManager, mRepo *mocks2.MockTestRepository)
		expectedErr  string
	}{
		{
			name:  "Success: entity delete",
			input: inputSuccess,
			prepareMocks: func(mTM *mocks.MockTransactionManager, mRepo *mocks2.MockTestRepository) {
				// эмулируем успешную транзакцию
				mTM.On("WithinTransaction", mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						fn := args.Get(2).(func(context.Context) error)
						_ = fn(ctx)
					})

				mRepo.On("Delete", mock.Anything, inputSuccess).Return(nil)
			},
			expectedErr: "",
		},
		{
			name:  "Error: delete failed ",
			input: inputFail,
			prepareMocks: func(mTM *mocks.MockTransactionManager, mRepo *mocks2.MockTestRepository) {
				mTM.On("WithinTransaction", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("db error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(2).(func(context.Context) error)
						_ = fn(ctx)
					})

				mRepo.On("Delete", mock.Anything, inputFail).Return(errors.New("some error"))
			},
			expectedErr: "delete test model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare
			mRepo := new(mocks2.MockTestRepository)
			mTM := new(mocks.MockTransactionManager)
			tt.prepareMocks(mTM, mRepo)
			uc := NewTestDeleteUseCase(mTM, mRepo)

			// act
			err := uc.Delete(ctx, tt.input)

			// assert
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			mRepo.AssertExpectations(t)
			mTM.AssertExpectations(t)
		})
	}
}
