package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/internal/domain"
	mocks2 "github.com/ElfAstAhe/go-service-template/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestListUseCase_List(t *testing.T) {
	// prepare
	expected := []*domain.Test{
		domain.NewTest("1", "1", "test 1", "", time.Now(), time.Now()),
		domain.NewTest("2", "2", "test 2", "", time.Now(), time.Now()),
		domain.NewTest("3", "3", "test 3", "", time.Now(), time.Now()),
	}
	ctx := context.Background()

	tests := []struct {
		name  string
		input struct {
			limit  int
			offset int
		}
		prepareMocks func(mRepo *mocks2.MockTestRepository)
		expectedRes  []*domain.Test
		expectedErr  string
	}{
		{
			name: "success 2 rows",
			input: struct {
				limit  int
				offset int
			}{
				limit:  2,
				offset: 0,
			},
			prepareMocks: func(mRepo *mocks2.MockTestRepository) {
				mRepo.On("List", mock.Anything, mock.Anything, mock.Anything).Return(expected[:2:2], nil)
			},
			expectedRes: expected[:2],
			expectedErr: "",
		},
		{
			name: "success all rows",
			input: struct {
				limit  int
				offset int
			}{
				limit:  10,
				offset: 0,
			},
			prepareMocks: func(mRepo *mocks2.MockTestRepository) {
				mRepo.On("List", mock.Anything, mock.Anything, mock.Anything).Return(expected, nil)
			},
			expectedRes: expected,
			expectedErr: "",
		},
		{
			name: "fail no rows",
			input: struct {
				limit  int
				offset int
			}{
				limit:  10,
				offset: 0,
			},
			prepareMocks: func(mRepo *mocks2.MockTestRepository) {
				mRepo.On("List", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db fail"))
			},
			expectedRes: nil,
			expectedErr: "list test data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			mRepo := new(mocks2.MockTestRepository)
			tt.prepareMocks(mRepo)

			uc := NewTestListUseCase(mRepo)

			// act
			res, err := uc.List(ctx, tt.input.limit, tt.input.offset)

			// assert
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.expectedRes, res)
			}

			mRepo.AssertExpectations(t)

		})
	}
}
