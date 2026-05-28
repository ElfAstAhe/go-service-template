package http

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/transport/http/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHTTPRunner_Lifecycle(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Infof", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Debug", mock.Anything).Return().Maybe()

	mockRouter := mocks2.NewMockRouter(t)
	mockRouter.On("GetRouter").Return(http.NewServeMux()) // Пустой хендлер

	conf := config.NewDefaultHTTPConfig()
	conf.Address = "127.0.0.1:0" // Свободный порт

	runner, err := NewRunner(
		WithConfig(conf),
		WithLogger("test-http", mockLog),
		WithRouter(mockRouter),
	)
	require.NoError(t, err)

	// Запускаем
	go func() {
		err := runner.Start(context.Background())
		assert.NoError(t, err, "Start should return nil when server is closed gracefully")
	}()

	// Ждем старта
	require.Eventually(t, func() bool { return runner.IsRunning() }, time.Second, 10*time.Millisecond)

	// Останавливаем
	err = runner.Stop(context.Background())
	assert.NoError(t, err)
	assert.False(t, runner.IsRunning())
}

func TestHTTPRunner_Start_IgnoreClosedError(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()

	mockRouter := mocks2.NewMockRouter(t)
	mockRouter.On("GetRouter").Return(http.NewServeMux())

	runner, _ := NewRunner(
		WithLogger("test", mockLog),
		WithRouter(mockRouter),
		WithServerLauncher(func(s *http.Server, c *config.HTTPConfig) error {
			return http.ErrServerClosed // Имитируем моментальный стоп
		}),
	)

	err := runner.Start(context.Background())
	assert.NoError(t, err, "Must ignore http.ErrServerClosed")
}
