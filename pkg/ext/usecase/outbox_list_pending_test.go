package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// OutboxListPendingTestSuite объединяет тесты для OutboxListPendingInteractor.
type OutboxListPendingTestSuite struct {
	suite.Suite
}

// TestOutboxListPendingTestSuite запускает сьют тестов.
func TestOutboxListPendingTestSuite(t *testing.T) {
	suite.Run(t, new(OutboxListPendingTestSuite))
}

// TestList проверяет успешные и ошибочные сценарии работы метода List.
func (suite *OutboxListPendingTestSuite) TestList() {
	ctx := context.Background()
	limit := 10

	var expectedOutboxes []domain.Outbox[string]
	repoError := errors.New("database connection failed")

	tests := []struct {
		name          string
		setupRepo     func(repo *mocks.MockOutboxRepository[domain.Outbox[string], string])
		expectedRes   []domain.Outbox[string]
		expectedErrFn func(t *testing.T, err error)
	}{
		{
			name: "Успешное получение списка сообщений outbox",
			setupRepo: func(repo *mocks.MockOutboxRepository[domain.Outbox[string], string]) {
				repo.On("ListPending", ctx, limit).
					Return(expectedOutboxes, nil).
					Once()
			},
			expectedRes: expectedOutboxes,
			expectedErrFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "Ошибка репозитория оборачивается в BllError",
			setupRepo: func(repo *mocks.MockOutboxRepository[domain.Outbox[string], string]) {
				repo.On("ListPending", ctx, limit).
					Return(nil, repoError).
					Once()
			},
			expectedRes: nil,
			expectedErrFn: func(t *testing.T, err error) {
				assert.Error(t, err)
				var bllErr *errs.BllError
				if assert.True(t, errors.As(err, &bllErr)) {
					assert.Contains(t, bllErr.Error(), "failed to list pending outboxes")
				}
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Инициализируем мок через автоматический конструктор Mockery нового поколения
			repoMock := mocks.NewMockOutboxRepository[domain.Outbox[string], string](suite.T())
			uc := NewOutboxListPending(repoMock)

			tt.setupRepo(repoMock)

			res, err := uc.List(ctx, limit)

			tt.expectedErrFn(suite.T(), err)
			assert.Equal(suite.T(), tt.expectedRes, res)
		})
	}
}
