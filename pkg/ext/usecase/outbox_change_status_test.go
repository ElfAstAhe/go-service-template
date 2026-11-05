package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/domain/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/ext/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// OutboxChangeStatusTestSuite объединяет тесты для OutboxChangeStatusInteractor.
type OutboxChangeStatusTestSuite struct {
	suite.Suite
}

// TestOutboxChangeStatusTestSuite запускает сьют тестов.
func TestOutboxChangeStatusTestSuite(t *testing.T) {
	suite.Run(t, new(OutboxChangeStatusTestSuite))
}

// TestChange проверяет валидацию, работу Unit of Work и обработку ошибок при смене статуса.
func (suite *OutboxChangeStatusTestSuite) TestChange() {
	ctx := context.Background()
	msgID := "test-msg-id"
	validStatus := "PENDING" // Подставь валидный статус из своего extdom
	repoError := errors.New("db error")

	tests := []struct {
		name          string
		id            string
		status        string
		setupMocks    func(uw *mocks.MockUnitOfWork, repo *mocks2.MockOutboxRepository[domain.Outbox[string], string])
		expectedErrFn func(t *testing.T, err error)
	}{
		{
			name:   "Ошибка валидации - пустой статус",
			id:     msgID,
			status: "   ",
			setupMocks: func(uw *mocks.MockUnitOfWork, repo *mocks2.MockOutboxRepository[domain.Outbox[string], string]) {
				// Метод должен завершиться до обращения к БД
			},
			expectedErrFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				var invalidArgErr *errs.InvalidArgumentError // Или твой тип ошибки валидации
				assert.True(t, errors.As(err, &invalidArgErr))
			},
		},
		{
			name:   "Успешное изменение статуса в транзакции Unit of Work",
			id:     msgID,
			status: validStatus,
			setupMocks: func(uw *mocks.MockUnitOfWork, repo *mocks2.MockOutboxRepository[domain.Outbox[string], string]) {
				// Мокаем выполнение Execute. Перехватываем функцию и запускаем её вручную.
				uw.On("Execute", ctx, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						// Получаем переданную функцию (второй аргумент)
						fn := args.Get(1).(func(context.Context) error)

						// Настраиваем поведение репозитория, который вызовется внутри этой функции
						repo.On("ChangeStatus", ctx, msgID, validStatus).
							Return(nil).
							Once()

						// Выполняем колбэк транзакции и возвращаем его результат наружу
						_ = fn(ctx)
					}).
					Once()
			},
			expectedErrFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "Ошибка репозитория внутри транзакции возвращает BllError",
			id:     msgID,
			status: validStatus,
			setupMocks: func(uw *mocks.MockUnitOfWork, repo *mocks2.MockOutboxRepository[domain.Outbox[string], string]) {
				uw.On("Execute", ctx, mock.Anything).
					Return(repoError). // Unit of Work возвращает ошибку, которую вернул репозиторий
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)

						repo.On("ChangeStatus", ctx, msgID, validStatus).
							Return(repoError).
							Once()

						_ = fn(ctx)
					}).
					Once()
			},
			expectedErrFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				var bllErr *errs.BllError
				if assert.True(t, errors.As(err, &bllErr)) {
					assert.Contains(t, bllErr.Error(), "failed")
				}
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Инициализируем новые чистые моки для каждого тест-кейса таблицы через MockNew конструкторы
			uwMock := mocks.NewMockUnitOfWork(suite.T())
			repoMock := mocks2.NewMockOutboxRepository[domain.Outbox[string], string](suite.T())

			uc := NewOutboxChangeStatus(uwMock, repoMock)

			tt.setupMocks(uwMock, repoMock)

			err := uc.Change(ctx, tt.id, tt.status)

			tt.expectedErrFn(suite.T(), err)
		})
	}
}
