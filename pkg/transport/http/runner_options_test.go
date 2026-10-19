package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/transport/http/mocks"
	"github.com/stretchr/testify/assert"
)

func TestFunctionalOptions_WithNewTypesAndMocks(t *testing.T) {
	t.Run("WithName корректно мутирует имя раннера", func(t *testing.T) {
		// Arrange
		r := &Runner{}
		expectedName := "user-service-http"

		// Act
		opt := WithName(expectedName)
		opt(r)

		// Assert
		assert.Equal(t, expectedName, r.name)
	})

	t.Run("WithRouter успешно инжектирует существующий мок роутера", func(t *testing.T) {
		// Arrange
		r := &Runner{}
		mockRouter := mocks2.NewMockRouter(t)

		// Act
		opt := WithRouter(mockRouter)
		opt(r)

		// Assert
		assert.Equal(t, mockRouter, r.router)
	})

	t.Run("WithConfig успешно инжектирует структуру конфигурации", func(t *testing.T) {
		// Arrange
		r := &Runner{}
		cfg := &config.HTTPConfig{
			ShutdownTimeout: 5 * time.Second,
		}

		// Act
		opt := WithConfig(cfg)
		opt(r)

		// Assert
		assert.Equal(t, cfg, r.conf)
	})

	t.Run("WithServerProvider успешно инжектирует провайдер с новой сигнатурой", func(t *testing.T) {
		// Arrange
		r := &Runner{}
		// Новая сигнатура: принимает router и conf, возвращает server и error
		var dummyProvider ServerProvider = func(router Router, conf *config.HTTPConfig) (*http.Server, error) {
			return &http.Server{}, nil
		}

		// Act
		opt := WithServerProvider(dummyProvider)
		opt(r)

		// Assert
		assert.NotNil(t, r.serverProvider)

		// Дополнительно проверяем, что сигнатура вызывается корректно
		srv, err := r.serverProvider(mocks2.NewMockRouter(t), &config.HTTPConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, srv)
	})

	t.Run("WithServerLauncher успешно инжектирует стратегию запуска с новой сигнатурой", func(t *testing.T) {
		// Arrange
		r := &Runner{}
		// Новая сигнатура: принимает server и conf, возвращает error
		var dummyLauncher ServerLauncher = func(server *http.Server, conf *config.HTTPConfig) error {
			return nil
		}

		// Act
		opt := WithServerLauncher(dummyLauncher)
		opt(r)

		// Assert
		assert.NotNil(t, r.serverLauncher)

		// Дополнительно проверяем вызов сигнатуры лаунчера
		err := r.serverLauncher(&http.Server{}, &config.HTTPConfig{})
		assert.NoError(t, err)
	})

	t.Run("WithShutdownTimeout меняет таймаут конфигурации", func(t *testing.T) {
		// Arrange
		r := &Runner{
			conf: &config.HTTPConfig{
				ShutdownTimeout: 2 * time.Second,
			},
		}
		newTimeout := 15 * time.Second

		// Act
		opt := WithShutdownTimeout(newTimeout)
		opt(r)

		// Assert
		assert.Equal(t, newTimeout, r.conf.ShutdownTimeout)
	})

	t.Run("WithLogger запрашивает дочерний логгер через существующий мок", func(t *testing.T) {
		// Arrange
		r := &Runner{}
		mockParentLogger := mocks.NewMockLogger(t)
		mockChildLogger := mocks.NewMockLogger(t)
		loggerName := "http_transport"

		// Контракт Mockery
		mockParentLogger.On("GetLogger", loggerName).Return(mockChildLogger).Once()

		// Act
		opt := WithLogger(loggerName, mockParentLogger)
		opt(r)

		// Assert
		assert.Equal(t, mockChildLogger, r.log)
		mockParentLogger.AssertExpectations(t)
	})
}
