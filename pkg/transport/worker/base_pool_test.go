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

func TestBasePool_SuccessLifecycle(t *testing.T) {
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return()

	taskProcessed := make(chan string, 2)
	jobHandler := func(ctx context.Context, workerIndex int, data string) error {
		taskProcessed <- data
		return nil
	}

	opts := &BasePoolOptions[string]{
		Name:            "test-pool",
		WorkerCount:     2,
		DataCapacity:    10,
		CompleteProcess: true,
		StopTimeout:     100 * time.Millisecond,
		Logger:          mockLogger,
		JobHandler:      jobHandler,
	}

	pool, err := NewBasePool[string](func(options *BasePoolOptions[string]) {
		*options = *opts
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pool.Start(ctx)
	require.NoError(t, err)

	pool.Push("task-1")
	assert.True(t, pool.TryPush("task-2"))

	var received []string
	for i := 0; i < 2; i++ {
		select {
		case task := <-taskProcessed:
			received = append(received, task)
		case <-time.After(50 * time.Millisecond):
			t.Fatal("timeout waiting for task execution")
		}
	}

	assert.Contains(t, received, "task-1")
	assert.Contains(t, received, "task-2")

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer stopCancel()

	err = pool.Stop(stopCtx)
	require.NoError(t, err)
}

func TestBasePool_ValidateFail(t *testing.T) {
	opts := &BasePoolOptions[string]{
		Name: "",
	}
	err := opts.Validate()
	assert.Error(t, err)
}
