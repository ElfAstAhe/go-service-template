package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	dommocks "github.com/ElfAstAhe/go-service-template/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestGetUseCase_Get(t *testing.T) {
	// prepare
	inputSuccess := "1"
	inputFail := "2"
	expected := domain.NewTest("1", "1", "test 1", "", time.Now(), time.Now())
	ctx := context.Background()

	tests := []struct {
		name         string
		input        string
		prepareMocks func(mRepo *dommocks.MockTestRepository)
		expectedRes  *domain.Test
		expectedErr  string
	}{
		{
			name:  "success",
			input: inputSuccess,
			prepareMocks: func(mRepo *dommocks.MockTestRepository) {
				mRepo.On("Find", mock.Anything, inputSuccess).Return(expected, nil)
			},
			expectedRes: expected,
			expectedErr: "",
		},
		{
			name:  "fail",
			input: inputFail,
			prepareMocks: func(mRepo *dommocks.MockTestRepository) {
				mRepo.On("Find", mock.Anything, inputFail).Return(nil, errors.New("some error"))
			},
			expectedRes: nil,
			expectedErr: "find test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare
			mRepo := new(dommocks.MockTestRepository)
			tt.prepareMocks(mRepo)
			uc := NewTestGetUseCase(mRepo)

			// act
			actual, err := uc.Get(ctx, tt.input)

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
		})
	}
}
