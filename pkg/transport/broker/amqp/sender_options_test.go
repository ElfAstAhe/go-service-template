package amqp

import (
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/broker/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSenderOptions_NewDefault_And_FluentAPI(t *testing.T) {
	// 1. Проверяем дефолты
	opts := NewSenderOptions()
	assert.NotNil(t, opts)
	assert.Equal(t, DefaultSenderConnectTimeout, opts.ConnectTimeout)
	assert.Equal(t, DefaultSenderPublishMaxTryAttempts, opts.PublishMaxTryAttempts)

	// 2. Проверяем Fluent API мутаторы
	mockConnector := mocks.NewMockConnector[*amqp.Session](t)
	mockLogger := mocks2.NewMockLogger(t)
	customSenderOpts := &amqp.SenderOptions{Name: "custom-sender"}
	customSendOpts := &amqp.SendOptions{Settled: false}

	WithSenderConnector(mockConnector)(opts)
	WithSenderTargetName("audit-topic")(opts)
	WithSenderConnectTimeout(10 * time.Second)(opts)
	WithSenderShutdownTimeout(12 * time.Second)(opts)
	WithSenderLogger(mockLogger)(opts)
	WithSenderPublishMaxTryAttempts(5)(opts)
	WithSenderPublishBaseRetryDelay(200 * time.Millisecond)(opts)
	WithSenderPublishMaxRetryDelay(5 * time.Second)(opts)
	WithSenderOpts(customSenderOpts)(opts)
	WithSendOpts(customSendOpts)(opts)

	assert.Equal(t, mockConnector, opts.Connector)
	assert.Equal(t, "audit-topic", opts.TargetName)
	assert.Equal(t, 10*time.Second, opts.ConnectTimeout)
	assert.Equal(t, 12*time.Second, opts.ShutdownTimeout)
	assert.Equal(t, mockLogger, opts.Logger)
	assert.Equal(t, 5, opts.PublishMaxTryAttempts)
	assert.Equal(t, 200*time.Millisecond, opts.PublishBaseRetryDelay)
	assert.Equal(t, 5*time.Second, opts.PublishMaxRetryDelay)
	assert.Equal(t, customSenderOpts, opts.Opts)
	assert.Equal(t, customSendOpts, opts.SendOpts)
}

func TestSenderOptions_Validate_Failures(t *testing.T) {
	// Базовая фабрика на 100% валидных опций Sender
	validBase := func(t *testing.T) *SenderOptions {
		return &SenderOptions{
			Connector:             mocks.NewMockConnector[*amqp.Session](t),
			TargetName:            "test-topic",
			ConnectTimeout:        5 * time.Second,
			ShutdownTimeout:       5 * time.Second,
			PublishMaxTryAttempts: 3,
			PublishBaseRetryDelay: 100 * time.Millisecond,
			PublishMaxRetryDelay:  2 * time.Second,
			Logger:                mocks2.NewMockLogger(t),
		}
	}

	tests := []struct {
		name           string
		mutate         func(so *SenderOptions)
		expectedPhrase string
	}{
		{
			name: "nil_connector",
			mutate: func(so *SenderOptions) {
				so.Connector = nil
			},
			expectedPhrase: "connector is required",
		},
		{
			name: "empty_target_name",
			mutate: func(so *SenderOptions) {
				so.TargetName = "   "
			},
			expectedPhrase: "target name empty",
		},
		{
			name: "invalid_connect_timeout",
			mutate: func(so *SenderOptions) {
				so.ConnectTimeout = 0
			},
			expectedPhrase: "connection timeout is invalid",
		},
		{
			name: "invalid_max_attempts",
			mutate: func(so *SenderOptions) {
				so.PublishMaxTryAttempts = -1
			},
			expectedPhrase: "publish max try attempts is invalid",
		},
		{
			name: "nil_logger",
			mutate: func(so *SenderOptions) {
				so.Logger = nil
			},
			expectedPhrase: "logger is nil",
		},
		{
			name: "base_retry_delay_greater_than_max",
			mutate: func(so *SenderOptions) {
				so.PublishBaseRetryDelay = 5 * time.Second
				so.PublishMaxRetryDelay = 1 * time.Second
			},
			expectedPhrase: "base retry delay cannot be greater",
		},
	}

	for _, tt := range tests {
		t.Run("sender_"+tt.name, func(t *testing.T) {
			opts := validBase(t)
			tt.mutate(opts)

			err := opts.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedPhrase)
		})
	}
}
