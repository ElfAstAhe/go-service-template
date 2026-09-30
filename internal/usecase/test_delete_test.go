package usecase

import (
	"context"
	"errors"
	"testing"

	dommocks "github.com/ElfAstAhe/go-service-template/internal/domain/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/domain/mocks"
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
		prepareMocks func(mUW *mocks2.MockUnitOfWork, mRepo *dommocks.MockTestRepository)
		expectedErr  string
	}{
		{
			name:  "Success: entity delete",
			input: inputSuccess,
			prepareMocks: func(mUW *mocks2.MockUnitOfWork, mRepo *dommocks.MockTestRepository) {
				// эмулируем успешную транзакцию
				mUW.On("Execute", mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(ctx)
					})

				mRepo.On("Delete", mock.Anything, inputSuccess).Return(nil)
			},
			expectedErr: "",
		},
		{
			name:  "Error: delete failed ",
			input: inputFail,
			prepareMocks: func(mUW *mocks2.MockUnitOfWork, mRepo *dommocks.MockTestRepository) {
				mUW.On("Execute", mock.Anything, mock.Anything).
					Return(errors.New("db error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
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
			mRepo := new(dommocks.MockTestRepository)
			mUW := new(mocks2.MockUnitOfWork)
			tt.prepareMocks(mUW, mRepo)
			uc := NewTestDeleteUseCase(mUW, mRepo)

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
			mUW.AssertExpectations(t)
		})
	}
}
