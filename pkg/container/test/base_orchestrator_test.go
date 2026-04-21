package test

import (
	"context"
	"errors"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/container"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/container/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBaseOrchestrator_FullLifecycle(t *testing.T) {
	// 1. Подготовка моков
	mockLog := mocks.NewMockLogger(t)
	// Настраиваем GetLogger, так как NewBaseOrchestrator его вызывает
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	orch := container.NewBaseOrchestrator(mockLog)
	ctx := context.Background()

	infraCtn := mocks2.NewMockContainer(t)
	appCtn := mocks2.NewMockContainer(t)

	// Настраиваем имена
	infraCtn.On("GetName").Return("infra")
	appCtn.On("GetName").Return("app")

	// Регистрация (проходит в основном потоке теста)
	require.NoError(t, orch.Register(infraCtn))
	require.NoError(t, orch.Register(appCtn))

	t.Run("Init_FIFO_Order", func(t *testing.T) {
		// Ожидаем логи (используем Anything для вариативной части, чтобы не возиться со слайсами)
		mockLog.On("Debugf", "initializing layer [%s]...", mock.Anything).Return().Once()
		infraCtn.On("Init", ctx).Return(nil).Once()

		mockLog.On("Debugf", "initializing layer [%s]...", mock.Anything).Return().Once()
		appCtn.On("Init", ctx).Return(nil).Once()

		err := orch.Init(ctx)
		assert.NoError(t, err)
	})

	t.Run("Close_LIFO_Order", func(t *testing.T) {
		// LIFO: Сначала закрывается app, потом infra
		mockLog.On("Debugf", "closing layer [%s]...", mock.Anything).Return().Once()
		appCtn.On("Close", ctx).Return(nil).Once()

		mockLog.On("Debugf", "closing layer [%s]...", mock.Anything).Return().Once()
		infraCtn.On("Close", ctx).Return(nil).Once()

		err := orch.Close(ctx)
		assert.NoError(t, err)
	})
}

func TestBaseOrchestrator_GetRunners(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	orch := container.NewBaseOrchestrator(mockLog)

	ctn := mocks2.NewMockContainer(t)
	run := mocks2.NewMockRunner(t)

	ctn.On("GetName").Return("transport")
	// Оркестратор запрашивает все имена в контейнере
	ctn.On("AllNames").Return([]string{"service", "server"}).Once()

	// Первый инстанс — просто объект (не Runner)
	ctn.On("GetInstance", "service").Return("just a string", nil).Once()
	// Второй инстанс — реализует Runner
	ctn.On("GetInstance", "server").Return(run, nil).Once()

	require.NoError(t, orch.Register(ctn))

	runners, err := orch.GetRunners()

	assert.NoError(t, err)
	assert.Len(t, runners, 1)
	assert.Equal(t, run, runners[0])
}

func TestBaseOrchestrator_Init_StopOnError(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	orch := container.NewBaseOrchestrator(mockLog)
	ctx := context.Background()

	c1 := mocks2.NewMockContainer(t)
	c2 := mocks2.NewMockContainer(t)

	c1.On("GetName").Return("c1")
	c2.On("GetName").Return("c2")

	// Первый контейнер возвращает ошибку
	initErr := errors.New("init failed")
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return().Once()
	c1.On("Init", ctx).Return(initErr).Once()

	// Второй контейнер НЕ должен вызываться вообще!

	require.NoError(t, orch.Register(c1))
	require.NoError(t, orch.Register(c2))

	err := orch.Init(ctx)
	assert.ErrorIs(t, err, initErr)

	// Проверяем, что методы c2 не вызывались
	c2.AssertNotCalled(t, "Init", ctx)
}

func TestBaseOrchestrator_GetContainer_NotFound(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	orch := container.NewBaseOrchestrator(mockLog)

	res, err := orch.GetContainer("unknown")
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
