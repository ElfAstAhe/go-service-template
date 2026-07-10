package azure

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/azure/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClientReceiver_Receive_Success_And_PayloadAssembly(t *testing.T) {
	// Arrange
	ctx := context.Background()
	queueName := "audit-events"

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", "azure-amqp-client-receiver").Return(mockLogger)

	// Имитируем успешное получение сообщения, разбитого на два чанка (Data [][]byte)
	mockAzureMsg := &amqp.Message{
		Data: [][]byte{
			[]byte("chunk-1_"),
			[]byte("chunk-2"),
		},
		ApplicationProperties: map[string]any{
			"trace_id": "uuid-12345",
		},
	}

	mockReceiver := new(mocks2.MockAmqpReceiverLink)
	mockReceiver.On("Receive", mock.Anything, mock.Anything).Return(mockAzureMsg, nil).Once()

	opts := NewClientReceiverOptions()
	opts.Logger = mockLogger

	cr, err := NewClientReceiver(func(cro *ClientReceiverOptions) {
		*cro = *opts
	})
	require.NoError(t, err)

	// Прогреваем кэш ресиверов моком
	cr.receivers[queueName] = mockReceiver

	// Act
	msg, err := cr.Receive(ctx, queueName, nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, msg)

	// Проверяем, что наша оптимизация через copy склеила Payload без искажений
	assert.Equal(t, []byte("chunk-1_chunk-2"), msg.Payload)
	assert.Equal(t, "uuid-12345", msg.Properties["trace_id"])
	assert.Equal(t, queueName, msg.TargetName) // Ресивер зафиксировал точный источник

	mockReceiver.AssertExpectations(t)
}

func TestClientReceiver_Accept_AddressRouting(t *testing.T) {
	// Arrange
	ctx := context.Background()
	queueName := "critical-audit-queue"

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)

	// Создаем фейковый оригинальный пакет amqp.Message
	origAzureMsg := &amqp.Message{}

	mockReceiver := new(mocks2.MockAmqpReceiverLink)
	// Проверяем, что AcceptMessage вызовется именно с нашей структурой пакета
	mockReceiver.On("AcceptMessage", mock.Anything, origAzureMsg).Return(nil).Once()

	opts := NewClientReceiverOptions()
	opts.Logger = mockLogger

	cr, err := NewClientReceiver(func(cro *ClientReceiverOptions) {
		*cro = *opts
	})
	require.NoError(t, err)

	// Кэшируем мок под конкретным именем очереди
	cr.receivers[queueName] = mockReceiver

	// Собираем конверт сообщения, пряча туда оригинал под ключом sysMsgKey
	msg := &pkgamqp.Message[*amqp.MessageHeader]{
		TargetName: queueName, // Указываем точную очередь-источник
		Properties: map[string]any{
			sysMsgKey: origAzureMsg,
		},
	}

	// Act
	err = cr.Accept(ctx, msg)

	// Assert
	// Метод должен отработать без ошибок, найдя по TargetName правильный линк в map
	assert.NoError(t, err)
	mockReceiver.AssertExpectations(t)
}

func TestClientReceiver_Receive_HandleReceiverFailure_InvalidatesCache(t *testing.T) {
	// Arrange
	ctx := context.Background()
	queueName := "flaky-queue"

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	// Настраиваем логгер на фиксацию падения линка
	mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return().Once()

	mockReceiver := new(mocks2.MockAmqpReceiverLink)
	// Симулируем критическую ошибку линка при вызове Receive
	linkErr := &amqp.LinkError{RemoteErr: &amqp.Error{Condition: amqp.ErrCondInternalError}}
	mockReceiver.On("Receive", mock.Anything, mock.Anything).Return(nil, linkErr).Once()

	opts := NewClientReceiverOptions()
	opts.Logger = mockLogger

	cr, err := NewClientReceiver(func(cro *ClientReceiverOptions) {
		*cro = *opts
	})
	require.NoError(t, err)

	cr.receivers[queueName] = mockReceiver

	// Act
	msg, err := cr.Receive(ctx, queueName, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.Contains(t, err.Error(), "azure receiver incoming packet error")

	// ПРОВЕРКА ИНВАЛИДАЦИИ: handleReceiverFailure обязан был удалить упавший топик из карты!
	cr.mu.RLock()
	_, exists := cr.receivers[queueName]
	cr.mu.RUnlock()
	assert.False(t, exists, "Битый линк ресивера должен быть атомарно удален из кэша пула")

	mockReceiver.AssertExpectations(t)
}

func TestClientReceiver_Close_ConcurrentClosing(t *testing.T) {
	// Arrange
	ctx := context.Background()

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything).Return().Maybe()

	mockReceiver1 := new(mocks2.MockAmqpReceiverLink)
	mockReceiver2 := new(mocks2.MockAmqpReceiverLink)

	// Имитируем долгое закрытие линков (сеть тормозит)
	mockReceiver1.On("Close", mock.Anything).Return(nil).After(15 * time.Millisecond)
	mockReceiver2.On("Close", mock.Anything).Return(nil).After(15 * time.Millisecond)

	opts := NewClientReceiverOptions()
	opts.Logger = mockLogger
	opts.ShutdownTimeout = 50 * time.Millisecond

	cr, err := NewClientReceiver(func(cro *ClientReceiverOptions) {
		*cro = *opts
	})
	require.NoError(t, err)

	cr.receivers["queue-1"] = mockReceiver1
	cr.receivers["queue-2"] = mockReceiver2

	start := time.Now()

	// Act
	err = cr.Close(ctx)

	duration := time.Since(start)

	// Assert
	assert.NoError(t, err)
	// Проверяем конкурентность: благодаря sync.WaitGroup и utils.NewConcurrentList,
	// оба зависших линка закроются одновременно, и Close уложится в ~15ms вместо последовательных 30ms.
	assert.Less(t, duration, 25*time.Millisecond, "Линки получателей закрываются последовательно")
	assert.Empty(t, cr.GetTargetNames())

	mockReceiver1.AssertExpectations(t)
	mockReceiver2.AssertExpectations(t)
}
