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

	// 1. Настраиваем мок логгера
	mockLogger := new(mocks.MockLogger) // Имя сгенерированного mockery мока для logger.Logger
	// Конструктор NewClientSingleSender вызывает GetLogger, настраиваем этот вызов
	mockLogger.On("GetLogger", "azure-amqp-client-single-sender").Return(mockLogger)

	// 2. Настраиваем DialFnTestGap на симуляцию падения сети
	mockDial := func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error) {
		return nil, expectedErr
	}

	// 3. Собираем опции
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

	// Пытаемся вызвать getSender, который внутри пойдет в getOrCreateSession -> mockDial
	sender, err := css.getSender(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, sender)

	// Проверяем, что наша кастомная обертка ошибок pkg/errs сохранила корневую причину
	assert.Contains(t, err.Error(), "dial failed")

	// Проверяем, что в процессе не случилось паники и структура осталась консистентной
	assert.Nil(t, css.connection)
	assert.Nil(t, css.session)

	mockLogger.AssertExpectations(t)
}

func TestClientSingleSender_WaitBackoff_And_ContextCancel(t *testing.T) {
	// Arrange
	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	// Метод waitBackoff пишет в Debugf, настраиваем мок на игнорирование или фиксацию этого вызова
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()

	opts := NewClientSenderOptions()
	opts.TargetName = "test-topic"
	opts.Logger = mockLogger
	opts.PublishBaseRetryDelay = 20 * time.Millisecond
	opts.PublishMaxRetryDelay = 100 * time.Millisecond

	css, err := NewClientSingleSender(func(cso *ClientSenderOptions) {
		*cso = *opts
	})
	require.NoError(t, err)

	// Тестируем быструю отмену контекста во время ожидания бэккоффа
	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем waitBackoff и отменяем контекст параллельно через небольшую паузу
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	start := time.Now()

	// Act
	// Передаем 3 попытку. Без отмены контекста код спал бы 20ms * 2^2 = 80ms (+/- джиттер)
	css.waitBackoff(ctx, 3)

	duration := time.Since(start)

	// Assert
	// Так как контекст отменился через 5ms, waitBackoff должен был мгновенно выйти,
	// не дожидаясь окончания всех 80ms.
	assert.Less(t, duration, 40*time.Millisecond, "waitBackoff не вышел досрочно при отмене контекста")

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
		opts.PublishMaxRetryDelay = 1 * time.Second // Базовая больше максимальной
		assert.Error(t, opts.Validate())
	})
}
