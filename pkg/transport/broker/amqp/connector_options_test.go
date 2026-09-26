package amqp

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
)

func TestConnectorOptions_NewDefault_And_FluentAPI(t *testing.T) {
	// 1. Проверяем дефолты
	opts := NewConnectorOptions()
	assert.NotNil(t, opts)
	assert.Equal(t, defaultConnectTimeout, opts.ConnectTimeout)
	assert.Equal(t, defaultShutdownTimeout, opts.ShutdownTimeout)

	// 2. Проверяем Fluent API мутаторы
	mockLogger := mocks.NewMockLogger(t)
	customConnOpts := &amqp.ConnOptions{ContainerID: "test-container"}
	customSessOpts := &amqp.SessionOptions{MaxLinks: 10}

	testDialFn := func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error) {
		return nil, nil
	}

	WithConnectorURL("amqps://localhost")(opts)
	WithConnectorConnectTimeout(8 * time.Second)(opts)
	WithConnectorShutdownTimeout(9 * time.Second)(opts)
	WithConnectorLogger(mockLogger)(opts)
	WithConnectorConnOpts(customConnOpts)(opts)
	WithConnectorSessionOpts(customSessOpts)(opts)
	WithConnectorDialFnTestGap(testDialFn)(opts)

	assert.Equal(t, "amqps://localhost", opts.URL)
	assert.Equal(t, 8*time.Second, opts.ConnectTimeout)
	assert.Equal(t, 9*time.Second, opts.ShutdownTimeout)
	assert.Equal(t, mockLogger, opts.Logger)
	assert.Equal(t, customConnOpts, opts.ConnOpts)
	assert.Equal(t, customSessOpts, opts.SessionOpts)
	assert.NotNil(t, opts.DialFnTestGap)
}

func TestConnectorOptions_Validate_Failures(t *testing.T) {
	// Базовая фабрика на 100% валидных опций Connector
	validBase := func(t *testing.T) *ConnectorOptions {
		return &ConnectorOptions{
			URL:             "amqps://sb-namespace.servicebus.windows.net",
			ConnectTimeout:  5 * time.Second,
			ShutdownTimeout: 5 * time.Second,
			Logger:          mocks.NewMockLogger(t),
		}
	}

	tests := []struct {
		name           string
		mutate         func(co *ConnectorOptions)
		expectedPhrase string
	}{
		{
			name: "empty_url",
			mutate: func(co *ConnectorOptions) {
				co.URL = "   "
			},
			expectedPhrase: "amqp connection URL cannot be empty",
		},
		{
			name: "nil_logger",
			mutate: func(co *ConnectorOptions) {
				co.Logger = nil
			},
			expectedPhrase: "connector logger is nil",
		},
		{
			name: "invalid_connect_timeout",
			mutate: func(co *ConnectorOptions) {
				co.ConnectTimeout = -1 * time.Second
			},
			expectedPhrase: "connection timeout is invalid",
		},
		{
			name: "invalid_shutdown_timeout",
			mutate: func(co *ConnectorOptions) {
				co.ShutdownTimeout = 0
			},
			expectedPhrase: "shutdown timeout is invalid",
		},
	}

	for _, tt := range tests {
		t.Run("connector_"+tt.name, func(t *testing.T) {
			opts := validBase(t)
			tt.mutate(opts)

			err := opts.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedPhrase)
		})
	}
}
