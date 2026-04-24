package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestGrpcRunner_Start_Success(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog).Maybe()
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Warnf", mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Infof", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Info", mock.Anything).Return().Maybe()

	conf := config.NewDefaultGRPCConfig()
	conf.Address = ":0" // Автоматический порт

	// Мокаем регистрацию сервиса
	regCalled := false
	regFunc := func(srv *grpc.Server) error {
		regCalled = true
		return nil
	}

	runner, err := NewRunner(
		WithConfig(conf),
		WithLogger("test", mockLog),
		WithServiceRegister(regFunc),
	)
	require.NoError(t, err)

	// Запускаем в горутине, так как Serve блокирует поток
	go func() {
		_ = runner.Start(context.Background())
	}()

	// Даем время на старт и проверяем статус
	assert.Eventually(t, func() bool { return runner.IsRunning() }, time.Second, 10*time.Millisecond)
	assert.True(t, regCalled, "ServiceRegister must be called during Start")

	// Чистим за собой
	_ = runner.Stop(context.Background())
}

func TestGrpcRunner_Start_FilteredError(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog).Maybe()
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Infof", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Info", mock.Anything).Return().Maybe()

	runner, _ := NewRunner(
		WithLogger("test", mockLog),
		WithServiceRegister(func(s *grpc.Server) error { return nil }),
		WithServerLauncher(func(s *grpc.Server, c *config.GRPCConfig) error {
			return grpc.ErrServerStopped // Имитируем остановку
		}),
	)

	// Эмулируем запуск. Должен вернуть nil, так как ошибка отфильтрована.
	err := runner.Start(context.Background())
	assert.NoError(t, err)
}
