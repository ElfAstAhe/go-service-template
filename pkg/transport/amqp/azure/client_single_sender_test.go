package azure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClientSingleSender_GetOrCreateSession_DialError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	expectedErr := errors.New("connection refused by target host")

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", "azure-amqp-client-single-sender").Return(mockLogger)

	mockDial := func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error) {
		return nil, expectedErr
	}

	opts := NewClientSenderOptions()
	opts.URL = "amqp://localhost:5672"
	opts.TargetName = "test-topic"
	opts.Logger = mockLogger
	opts.DialFnTestGap = mockDial

	// Act
	css, err := NewClientSingleSender(func(cso *ClientSenderOptions) {
		*cso = *opts
	})
	require.NoError(t, err)

	sender, err := css.getSender(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, sender)
	assert.Contains(t, err.Error(), "dial failed")
	assert.Nil(t, css.connection)
	assert.Nil(t, css.session)

	mockLogger.AssertExpectations(t)
}

func TestClientSingleSender_WaitBackoff_And_ContextCancel(t *testing.T) {
	// Arrange
	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)

	// Оптимизация: Разрешаем вызывать логгеру Debugf любое количество раз, чтобы тесты не были хрупкими
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()

	opts := NewClientSenderOptions()
	opts.TargetName = "test-topic"
	opts.Logger = mockLogger
	// Ставим задержки чуть больше, чтобы разница между штатным сном и прерыванием была очевидной
	opts.PublishBaseRetryDelay = 100 * time.Millisecond
	opts.PublishMaxRetryDelay = 500 * time.Millisecond

	css, err := NewClientSingleSender(func(cso *ClientSenderOptions) {
		*cso = *opts
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Избавляемся от flaky-поведения: отменяем контекст сразу же в фоновом потоке.
	// Канал syncChan гарантирует, что горутина запустилась.
	syncChan := make(chan struct{})
	go func() {
		close(syncChan)
		cancel()
	}()
	<-syncChan

	start := time.Now()

	// Act
	// Попытка 3 при базовой 100ms должна спать 400ms. Но контекст уже отменен.
	css.waitBackoff(ctx, 3)

	duration := time.Since(start)

	// Assert
	// Выход должен быть мгновенным (явно меньше 400ms и даже меньше 50ms)
	assert.Less(t, duration, 50*time.Millisecond, "waitBackoff не вышел досрочно при отмене контекста")

	mockLogger.AssertExpectations(t)
}

func TestClientSenderOptions_ValidationCases(t *testing.T) {
	mockLogger := new(mocks.MockLogger)

	t.Run("empty URL", func(t *testing.T) {
		opts := NewClientSenderOptions()
		opts.URL = ""
		opts.TargetName = "queue"
		opts.Logger = mockLogger
		assert.Error(t, opts.Validate())
	})

	t.Run("empty TargetName", func(t *testing.T) {
		opts := NewClientSenderOptions()
		opts.URL = "amqp://localhost"
		opts.TargetName = ""
		opts.Logger = mockLogger
		assert.Error(t, opts.Validate())
	})

	t.Run("nil Logger", func(t *testing.T) {
		opts := NewClientSenderOptions()
		opts.URL = "amqp://localhost"
		opts.TargetName = "queue"
		opts.Logger = nil
		assert.Error(t, opts.Validate())
	})

	t.Run("invalid delays", func(t *testing.T) {
		opts := NewClientSenderOptions()
		opts.URL = "amqp://localhost"
		opts.TargetName = "queue"
		opts.Logger = mockLogger
		opts.PublishBaseRetryDelay = 5 * time.Second
		opts.PublishMaxRetryDelay = 1 * time.Second
		assert.Error(t, opts.Validate())
	})
}
