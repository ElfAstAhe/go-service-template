package test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/container/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBaseContainer_Lifecycle(t *testing.T) {
	mockOrchestrator := mocks.NewMockOrchestrator(t)
	mockLog := mocks2.NewMockLogger(t)
	// Настраиваем GetLogger, так как NewBaseOrchestrator его вызывает
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	c := container.NewBaseContainer(
		container.WithName("test-container"),
		container.WithOrchestrator(mockOrchestrator),
		container.WithLogger(mockLog))

	t.Run("Add_And_Get_Success", func(t *testing.T) {
		instance := "hello-world"
		err := c.RegisterInstance("key1", instance)

		assert.NoError(t, err)
		res, err := c.GetInstance("key1")
		assert.NoError(t, err)
		assert.Equal(t, instance, res)
	})

	t.Run("Add_Duplicate_Error", func(t *testing.T) {
		err := c.RegisterInstance("key1", "duplicate")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Get_NonExistent_Error", func(t *testing.T) {
		res, err := c.GetInstance("unknown")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "not registered")
	})

	t.Run("Remove_Instance", func(t *testing.T) {
		_ = c.RegisterInstance("to-delete", 123)
		err := c.UnregisterInstance("to-delete")
		assert.NoError(t, err)

		_, err = c.GetInstance("to-delete")
		assert.Error(t, err)
	})

	t.Run("Validate_Empty_Name", func(t *testing.T) {
		err := c.RegisterInstance("", "something")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is empty")
	})
}

func TestBaseContainer_Concurrency(t *testing.T) {
	mockLog := mocks2.NewMockLogger(t)
	// Настраиваем GetLogger, так как NewBaseOrchestrator его вызывает
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	c := container.NewBaseContainer(
		container.WithName("concurrency-test"),
		container.WithLogger(mockLog))

	const workers = 30
	const iterations = 50
	wg := sync.WaitGroup{}
	wg.Add(workers * 2)

	// Параллельная запись
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("w-%d-k-%d", workerID, j)
				_ = c.RegisterInstance(key, j)
			}
		}(i)
	}

	// Параллельное чтение (ошибки отсутствия ключа здесь допустимы и безопасны)
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("w-%d-k-%d", workerID, j)
				_, _ = c.GetInstance(key)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, workers*iterations, len(c.AllNames()))
}

func TestBaseContainer_Close(t *testing.T) {
	t.Run("Close_Success_With_All_Interfaces", func(t *testing.T) {
		mockLog := mocks2.NewMockLogger(t)
		// Настраиваем GetLogger, так как NewBaseOrchestrator его вызывает
		mockLog.On("GetLogger", mock.Anything).Return(mockLog)
		c := container.NewBaseContainer(
			container.WithName("close-success"),
			container.WithLogger(mockLog))

		// Создаем типизированные моки из пакета mocks
		mockCloser := mocks.NewMockSimpleCloser(t)
		mockCloser.On("Close").Return(nil).Once()

		mockCtxCloser := mocks.NewMockContextCloser(t)
		mockCtxCloser.On("Close", mock.Anything).Return(nil).Once()

		require.NoError(t, c.RegisterInstance("io-closer", mockCloser))
		require.NoError(t, c.RegisterInstance("ctx-closer", mockCtxCloser))
		require.NoError(t, c.RegisterInstance("string-instance", "non-closable"))

		err := c.Close(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 0, len(c.AllNames())) // Мапа должна полностью очиститься

		mockCloser.AssertExpectations(t)
		mockCtxCloser.AssertExpectations(t)
	})

	t.Run("Close_With_Errors_Returns_Combined_Error", func(t *testing.T) {
		mockLog := mocks2.NewMockLogger(t)
		// Настраиваем GetLogger, так как NewBaseOrchestrator его вызывает
		mockLog.On("GetLogger", mock.Anything).Return(mockLog)
		c := container.NewBaseContainer(
			container.WithName("close-errors"),
			container.WithLogger(mockLog))

		errCloser := errors.New("io closer failed")
		errCtxCloser := errors.New("context closer failed")

		mockCloser := mocks.NewMockSimpleCloser(t)
		mockCloser.On("Close").Return(errCloser).Once()

		mockCtxCloser := mocks.NewMockContextCloser(t)
		mockCtxCloser.On("Close", mock.Anything).Return(errCtxCloser).Once()

		require.NoError(t, c.RegisterInstance("closer", mockCloser))
		require.NoError(t, c.RegisterInstance("ctxCloser", mockCtxCloser))

		err := c.Close(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "io closer failed")
		assert.Contains(t, err.Error(), "context closer failed")
	})

	t.Run("Close_Timeout_Limit_Reached", func(t *testing.T) {
		mockLog := mocks2.NewMockLogger(t)
		// Настраиваем GetLogger, так как NewBaseOrchestrator его вызывает
		mockLog.On("GetLogger", mock.Anything).Return(mockLog)
		c := container.NewBaseContainer(
			container.WithName("close-timeout"),
			container.WithLogger(mockLog))

		mockCtxCloser := mocks.NewMockContextCloser(t)
		// Имитируем зависание ресурса, заставляя мок ждать отмены контекста
		mockCtxCloser.On("Close", mock.Anything).Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			<-ctx.Done()
		}).Return(context.DeadlineExceeded).Once()

		require.NoError(t, c.RegisterInstance("hanging-service", mockCtxCloser))

		// Выделяем жесткий лимит времени на выполнение Close
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := c.Close(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "close timeout limit reached")
	})
}
