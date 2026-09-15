package kafka

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/kafka/mocks"
	mocks3 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/mocks"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReceiver_Receive_Success_And_MemoryAllocation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaReceiverLink(t)
	mockConnector := mocks3.NewMockConnector[*kafka.Client](t)

	mockLogger := mocks.NewMockLogger(t)
	// Защищаем логгер от строгих проверок вызовов, так как на Fast Path он не пишется
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger).Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	// Сымитируем пакет из Kafka
	expectedKafkaMsg := kafka.Message{
		Topic: "tiny.auth.login.attempts",
		Key:   []byte("user_test"),
		Value: []byte(`{"status":"fail"}`),
	}
	mockLink.On("FetchMessage", ctx).Return(expectedKafkaMsg, nil)

	receiver := &Receiver{
		opts: &ReceiverOptions{
			Connector:  mockConnector,
			TargetName: "tiny.auth.login.attempts",
			GroupID:    "auth-group",
		},
		reader: mockLink,
		logger: mockLogger,
	}

	// Act
	msg, err := receiver.Receive(ctx, nil)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, `{"status":"fail"}`, string(msg.GetPayload()))
	assert.Equal(t, "user_test", msg.GetProperties()["kafka_message_key"])

	// ПРОВЕРКА ESCAPE-АНАЛИЗА И ЗАЩИТЫ ОТ МУТАЦИЙ ПАМЯТИ
	rawOriginal, err := msg.ExtractOriginalMessage()
	require.NoError(t, err)

	extractedMsg, ok := rawOriginal.(*kafka.Message)
	require.True(t, ok)
	assert.Equal(t, expectedKafkaMsg.Topic, extractedMsg.Topic)
	assert.False(t, &expectedKafkaMsg == extractedMsg, "Critical vulnerability: extracted message leaks iterator memory pointer")

	mockLink.AssertExpectations(t)
}

func TestReceiver_Accept_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaReceiverLink(t)
	mockConnector := mocks3.NewMockConnector[*kafka.Client](t)

	// Создаем наше упакованное сообщение, имитируя, что оно пришло из Receive()
	kafkaMsg := kafka.Message{
		Topic:     "tiny.auth.login.attempts",
		Partition: 1,
		Offset:    42,
	}

	transportMsg := &Message{
		TargetName: "tiny.auth.login.attempts",
		Payload:    []byte(`{}`),
		Props: map[string]any{
			sysMsgKey: &kafkaMsg,
		},
	}

	// Ожидаем вариативный вызов CommitMessages (используем mock.Anything для бронебойности с вариативностью Mockery)
	mockLink.On("CommitMessages", ctx, mock.Anything).Return(nil)

	receiver := &Receiver{
		opts: &ReceiverOptions{
			Connector:  mockConnector,
			TargetName: "tiny.auth.login.attempts",
		},
		reader: mockLink,
		logger: mocks.NewMockLogger(t),
	}

	// Act
	err := receiver.Accept(ctx, transportMsg)

	// Assert
	assert.NoError(t, err)
	mockLink.AssertExpectations(t)
}

func TestReceiver_HandleReceiverFailure_InvalidatesConnector(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaReceiverLink(t)
	mockConnector := mocks3.NewMockConnector[*kafka.Client](t)

	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger).Maybe()
	mockLogger.On("Warnf", mock.Anything, mock.Anything).Return().Maybe()

	// Настраиваем падение по сетевой ошибке
	mockNetErr := &net.OpError{Op: "read", Net: "tcp", Err: errors.New("connection reset")}
	mockLink.On("FetchMessage", ctx).Return(kafka.Message{}, mockNetErr)

	// Ожидаем асинхронное закрытие старого линка в фоне через go-рутину внутри handleReceiverFailure
	mockLink.On("Close").Return(nil).Maybe()

	// Ожидаем, что ресивер обязан вызвать Invalidate у глобального коннектора
	mockConnector.On("Invalidate", mockNetErr).Return()

	receiver := &Receiver{
		opts: &ReceiverOptions{
			Connector:  mockConnector,
			TargetName: "tiny.auth.login.attempts",
		},
		reader: mockLink,
		logger: mockLogger,
	}

	// Act
	_, err := receiver.Receive(ctx, nil)

	// Assert
	assert.Error(t, err)

	// Даем горутине завершить асинхронный сброс ридера
	time.Sleep(10 * time.Millisecond)

	receiver.mu.RLock()
	assert.Nil(t, receiver.reader, "Receiver link was not reset to nil after failure")
	receiver.mu.RUnlock()

	mockConnector.AssertExpectations(t)
}
