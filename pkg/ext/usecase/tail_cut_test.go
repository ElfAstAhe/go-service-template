package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/domain/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/ext/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTailCutInteractor_Cut_WithTestify(t *testing.T) {
	targetID := "db050a41-8651-4e76-88a4-0980c651f644"
	errDb := errors.New("postgres: row is locked or connection timeout")

	tests := []struct {
		name    string
		setup   func(uw *mocks.MockUnitOfWork, repo *mocks2.MockTailRepository[string])
		wantErr bool
	}{
		{
			name: "Успешное попоточное удаление записи Outbox в рамках ACID транзакции",
			setup: func(uw *mocks.MockUnitOfWork, repo *mocks2.MockTailRepository[string]) {
				// Динамически вычисляем возвращаемое значение: запускаем колбэк и транслируем его результат наружу
				uw.On("Execute", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.On("Delete", mock.Anything, targetID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Сбой операции: репозиторий возвращает ошибку, интерактор упаковывает её в BllError",
			setup: func(uw *mocks.MockUnitOfWork, repo *mocks2.MockTailRepository[string]) {
				// ТРАНСЛЯЦИЯ ОШИБКИ: Теперь UnitOfWork честно вернет то, что выбросит репозиторий внутри транзакции
				uw.On("Execute", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.On("Delete", mock.Anything, targetID).Return(errDb).Once()
			},
			wantErr: true,
		},
		{
			name: "Сбой операции: сам компонент UnitOfWork возвращает ошибку до вызова репозитория",
			setup: func(uw *mocks.MockUnitOfWork, repo *mocks2.MockTailRepository[string]) {
				// Имитируем сбой самого транзактора (например, пул соединений пуст), колбэк fn даже не вызывается
				uw.On("Execute", mock.Anything, mock.Anything).Return(errors.New("tx open failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockUW := mocks.NewMockUnitOfWork(t)
			mockRepo := mocks2.NewMockTailRepository[string](t)

			tt.setup(mockUW, mockRepo)

			interactor := NewTailCutUseCase[string](mockUW, mockRepo)
			err := interactor.Cut(ctx, targetID)

			if tt.wantErr {
				assert.Error(t, err)

				var bllErr *errs.BllError
				assert.ErrorAs(t, err, &bllErr, "Ожидалось, что ошибка будет упакована в *errs.BllError")
			} else {
				assert.NoError(t, err)
			}

			mockUW.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}
