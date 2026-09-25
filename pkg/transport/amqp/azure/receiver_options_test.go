package azure

import (
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/mocks"
	"github.com/stretchr/testify/assert"
)

func TestReceiverOptions_NewDefault_And_FluentAPI(t *testing.T) {
	// 1. Проверяем дефолты
	opts := NewReceiverOptions()
	assert.NotNil(t, opts)
	assert.Equal(t, DefaultReceiverConnectTimeout, opts.ConnectTimeout)
	assert.Equal(t, DefaultReceiverLinkCredit, opts.LinkCredit)

	// 2. Проверяем Fluent API мутаторы
	mockConnector := mocks.NewMockConnector[*amqp.Session](t)
	mockLogger := mocks2.NewMockLogger(t)
	customReceiverOpts := &amqp.ReceiverOptions{Credit: 50}
	customReceiveOpts := &amqp.ReceiveOptions{}

	WithReceiverConnector(mockConnector)(opts)
	WithReceiverTargetName("audit-queue")(opts)
	WithReceiverConnectTimeout(4 * time.Second)(opts)
	WithReceiverShutdownTimeout(6 * time.Second)(opts)
	WithReceiverLogger(mockLogger)(opts)
	WithReceiverLinkCredit(200)(opts)
	WithReceiverOpts(customReceiverOpts)(opts)
	WithReceiverReceiveOpts(customReceiveOpts)(opts)

	assert.Equal(t, mockConnector, opts.Connector)
	assert.Equal(t, "audit-queue", opts.TargetName)
	assert.Equal(t, 4*time.Second, opts.ConnectTimeout)
	assert.Equal(t, 6*time.Second, opts.ShutdownTimeout)
	assert.Equal(t, mockLogger, opts.Logger)
	assert.Equal(t, int32(200), opts.LinkCredit)
	assert.Equal(t, customReceiverOpts, opts.ReceiverOpts)
	assert.Equal(t, customReceiveOpts, opts.ReceiveOpts)
}

func TestReceiverOptions_Validate_Failures(t *testing.T) {
	// Базовая фабрика на 100% валидных опций Receiver
	validBase := func(t *testing.T) *ReceiverOptions {
		return &ReceiverOptions{
			Connector:       mocks.NewMockConnector[*amqp.Session](t),
			TargetName:      "test-queue",
			ConnectTimeout:  5 * time.Second,
			ShutdownTimeout: 5 * time.Second,
			LinkCredit:      100,
			Logger:          mocks2.NewMockLogger(t),
		}
	}

	tests := []struct {
		name           string
		mutate         func(ro *ReceiverOptions)
		expectedPhrase string
	}{
		{
			name: "nil_connector",
			mutate: func(ro *ReceiverOptions) {
				ro.Connector = nil
			},
			expectedPhrase: "connector is required",
		},
		{
			name: "empty_target_name",
			mutate: func(ro *ReceiverOptions) {
				ro.TargetName = ""
			},
			expectedPhrase: "target name (queue/topic) cannot be empty",
		},
		{
			name: "invalid_shutdown_timeout",
			mutate: func(ro *ReceiverOptions) {
				ro.ShutdownTimeout = -2 * time.Second
			},
			expectedPhrase: "shutdown timeout is invalid",
		},
		{
			name: "zero_link_credit",
			mutate: func(ro *ReceiverOptions) {
				ro.LinkCredit = 0
			},
			expectedPhrase: "link credit must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run("receiver_"+tt.name, func(t *testing.T) {
			opts := validBase(t)
			tt.mutate(opts)

			err := opts.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedPhrase)
		})
	}
}
