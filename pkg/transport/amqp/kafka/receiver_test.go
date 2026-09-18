package kafka

import (
	"context"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/kafka/mocks"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReceiver_Receive_Success_And_MemoryAllocation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaReceiverLink(t)
	mockLogger := mocks.NewMockLogger(t)

	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger).Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	// Сымитируем пакет из Kafka с заголовком ключа
	expectedKafkaMsg := kafka.Message{
		Topic: "tiny.auth.login.attempts",
		Key:   []byte("user_test"),
		Value: []byte(`{"status":"fail"}`),
	}
	mockLink.On("FetchMessage", ctx).Return(expectedKafkaMsg, nil)

	receiver := &Receiver{
		opts: &ReceiverOptions{
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

	// Проверка уникальности адреса (адрес выделенной структуры в NewMessage не равен адресу исходной на стеке теста)
	assert.False(t, &expectedKafkaMsg == extractedMsg, "Critical vulnerability: extracted message leaks iterator memory pointer")

	mockLink.AssertExpectations(t)
}

func TestReceiver_Accept_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaReceiverLink(t)

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
			sysKafkaMsgKey: &kafkaMsg,
		},
	}

	// Ожидаем вариативный вызов CommitMessages
	mockLink.On("CommitMessages", ctx, mock.Anything).Return(nil)

	receiver := &Receiver{
		opts: &ReceiverOptions{
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
