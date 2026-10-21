package tailcutter

import (
	"context"
	"errors"
	"testing"
	"time"

	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/ext/usecase/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorker_DataProvider(t *testing.T) {
	t.Run("Успешное получение данных, когда очистка включена (Happy Path)", func(t *testing.T) {
		// Arrange
		mockLogger := mocks.NewMockLogger(t)
		mockGetUC := mocks2.NewMockTailGetUseCase[string](t)
		mockCutUC := mocks2.NewMockTailCutUseCase[string](t)

		// Настраиваем логгер: методы Debugf вызываются на предсохраненном логгере tc.logger
		mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()

		opts := &Options{
			PeriodicDispatcherOptions: &worker.PeriodicDispatcherOptions[string]{
				Name:   "test",
				Logger: mockLogger,
			},
			CutEnabled:   true,
			DataInterval: 5 * time.Minute,
		}

		expectedData := []string{"id-1", "id-2", "id-3"}
		ctx := context.Background()
		eventTime := time.Now()
		expectedCutTime := eventTime.Add(-opts.DataInterval)

		mockGetUC.On("GetTail", ctx, expectedCutTime).Return(expectedData, nil).Once()

		tc := &Worker{
			name:      "tail-cutter-test",
			opts:      opts,
			tailGetUC: mockGetUC,
			tailCutUC: mockCutUC,
			logger:    mockLogger, // Прокидываем мок как готовый изолированный логгер
		}

		// Act
		res, err := tc.dataProvider(ctx, eventTime)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedData, res)
	})

	t.Run("Возврат пустого слайса, если очистка выключена (CutEnabled = false)", func(t *testing.T) {
		// Arrange
		mockLogger := mocks.NewMockLogger(t)
		mockGetUC := mocks2.NewMockTailGetUseCase[string](t)

		mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()

		opts := &Options{
			PeriodicDispatcherOptions: &worker.PeriodicDispatcherOptions[string]{
				Name:   "test",
				Logger: mockLogger,
			},
			CutEnabled:   false,
			DataInterval: 5 * time.Minute,
		}

		tc := &Worker{
			name:      "tail-cutter-test",
			opts:      opts,
			tailGetUC: mockGetUC,
			logger:    mockLogger,
		}

		// Act
		res, err := tc.dataProvider(context.Background(), time.Now())

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, res)
		mockGetUC.AssertNotCalled(t, "GetTail", mock.Anything, mock.Anything)
	})

	t.Run("Трансляция ошибки доменного UseCase наверх", func(t *testing.T) {
		// Arrange
		mockLogger := mocks.NewMockLogger(t)
		mockGetUC := mocks2.NewMockTailGetUseCase[string](t)

		mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()

		opts := &Options{
			PeriodicDispatcherOptions: &worker.PeriodicDispatcherOptions[string]{
				Name:   "test",
				Logger: mockLogger,
			},
			CutEnabled:   true,
			DataInterval: 5 * time.Minute,
		}

		expectedErr := errors.New("db connection failure")
		mockGetUC.On("GetTail", mock.Anything, mock.Anything).Return(nil, expectedErr).Once()

		tc := &Worker{
			opts:      opts,
			tailGetUC: mockGetUC,
			logger:    mockLogger,
		}

		// Act
		res, err := tc.dataProvider(context.Background(), time.Now())

		// Assert
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, res)
	})
}
