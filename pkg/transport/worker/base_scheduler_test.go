package worker

import (
	"context"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBaseScheduler_SuccessLifecycle(t *testing.T) {
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()

	dispatcherCalled := make(chan time.Time, 1)
	timerDispatcher := func(ctx context.Context, eventTime time.Time) error {
		dispatcherCalled <- eventTime
		return nil
	}

	opts := &BaseSchedulerOptions{
		Name:             "test-cleaner-scheduler",
		StartInterval:    5 * time.Millisecond,
		ScheduleInterval: 20 * time.Millisecond,
		StopTimeout:      50 * time.Millisecond,
		Logger:           mockLogger,
		TimerDispatcher:  timerDispatcher,
	}

	scheduler, err := NewBaseScheduler(func(options *BaseSchedulerOptions) {
		*options = *opts
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-dispatcherCalled:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for scheduler tick event")
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer stopCancel()

	err = scheduler.Stop(stopCtx)
	require.NoError(t, err)
}

func TestBaseScheduler_ValidateFail(t *testing.T) {
	opts := &BaseSchedulerOptions{
		Name: "",
	}
	err := opts.Validate()
	assert.Error(t, err)
}
