package worker

import (
	"context"
	"testing"
	"time"

	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDispatcherDataProvider[D comparable] struct{ mock.Mock }

func (m *MockDispatcherDataProvider[D]) Execute(ctx context.Context, t time.Time) ([]D, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]D), args.Error(1)
}

func TestBaseSchedulerDispatcher_SuccessLifecycle(t *testing.T) {
	mockProvider := new(MockDispatcherDataProvider[string])
	mockPool := mocks.NewMockPool[string](t)
	mockLogger := mocks2.NewMockLogger(t)

	expectedRecords := []string{"audit-log-1", "audit-log-2"}
	targetTime := time.Now()

	mockProvider.On("Execute", mock.Anything, mock.AnythingOfType("time.Time")).Return(expectedRecords, nil).Once()
	mockPool.On("Push", "audit-log-1").Return().Once()
	mockPool.On("Push", "audit-log-2").Return().Once()

	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()

	opts := &BaseSchedulerDispatcherOptions[string]{
		Name:             "test-scheduler-dispatcher",
		WorkerCount:      2,
		DataCapacity:     10,
		CompleteProcess:  true,
		StartInterval:    10 * time.Millisecond,
		ScheduleInterval: 50 * time.Millisecond,
		StopTimeout:      100 * time.Millisecond,
		DataProvider:     mockProvider.Execute,
		Logger:           mockLogger,
	}

	dispatcher := &BaseSchedulerDispatcher[string]{
		name:         opts.Name,
		workerPool:   mockPool,
		dataProvider: opts.DataProvider,
		opts:         opts,
		log:          mockLogger,
	}

	ctx := context.Background()
	err := dispatcher.timerDispatcher(ctx, targetTime)

	require.NoError(t, err, "timerDispatcher execution loop must complete without unexpected runtime faults")
	mockProvider.AssertExpectations(t)
	mockPool.AssertExpectations(t)
}

func TestBaseSchedulerDispatcher_ValidateFail(t *testing.T) {
	opts := &BaseSchedulerDispatcherOptions[string]{
		Name: "",
	}

	err := opts.Validate()
	assert.Error(t, err, "validation phase must explicitly return a configuration mapping tracking failure descriptor")
}
