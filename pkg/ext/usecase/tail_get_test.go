package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/ext/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTailGetInteractor_GetTail_WithTestify(t *testing.T) {
	mockTime := time.Now().Add(-1 * time.Hour)
	expectedIDs := []string{"id-1", "id-2", "id-3"}

	tests := []struct {
		name    string
		tail    time.Time
		setup   func(repo *mocks2.MockTailRepository[string])
		wantErr bool
		errType any
	}{
		{
			name: "Успешное извлечение среза идентификаторов по валидной временной метке",
			tail: mockTime,
			setup: func(repo *mocks2.MockTailRepository[string]) {
				repo.On("GetTail", mock.Anything, mockTime).Return(expectedIDs, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Сбой операции: передана нулевая временная метка, срабатывает Fail-Fast валидация",
			tail: time.Time{}, // IsZero() == true
			setup: func(repo *mocks2.MockTailRepository[string]) {
				// Репозиторий вообще не должен вызываться
			},
			wantErr: true,
			errType: &errs.InvalidArgumentError{},
		},
		{
			name: "Сбой операции: репозиторий возвращает инфраструктурную ошибку базы данных",
			tail: mockTime,
			setup: func(repo *mocks2.MockTailRepository[string]) {
				repo.On("GetTail", mock.Anything, mockTime).Return(nil, errors.New("db disconnect")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := mocks2.NewMockTailRepository[string](t)

			tt.setup(mockRepo)

			interactor := NewTailGetUseCase[string](mockRepo)
			res, err := interactor.GetTail(ctx, tt.tail)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
				if tt.errType != nil {
					// Проверяем точечный тип ошибки валидации аргументов
					var argErr *errs.InvalidArgumentError
					assert.ErrorAs(t, err, &argErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedIDs, res)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
