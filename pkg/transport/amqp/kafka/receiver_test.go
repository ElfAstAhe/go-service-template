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
			Partition:  -1, // Дефолт для режима группы
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

// 3. НОВЫЙ ТЕСТ: Успешное чтение в режиме Direct Consumer (Без GroupID, по конкретной партиции)
func TestReceiver_Receive_Standalone_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaReceiverLink(t)
	mockLogger := mocks.NewMockLogger(t)

	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger).Maybe()

	expectedKafkaMsg := kafka.Message{
		Topic:     "tiny.auth.login.attempts",
		Partition: 3, // Читаем конкретную 3-ю партицию
		Value:     []byte(`{"status":"success"}`),
	}
	mockLink.On("FetchMessage", ctx).Return(expectedKafkaMsg, nil)

	receiver := &Receiver{
		opts: &ReceiverOptions{
			TargetName: "tiny.auth.login.attempts",
			GroupID:    "", // Пустая группа включает Standalone режим
			Partition:  3,
		},
		reader: mockLink,
		logger: mockLogger,
	}

	// Act
	msg, err := receiver.Receive(ctx, nil)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, `{"status":"success"}`, string(msg.GetPayload()))

	mockLink.AssertExpectations(t)
}

// TestReceiver_Stats_Success проверяет маппинг рантайм-статистики методом Stats()
// и корректность заполнения блоков COMMON и KAFKA.
func TestReceiver_Stats_Success(t *testing.T) {
	// Arrange
	mockLink := mocks2.NewMockKafkaReceiverLink(t)

	// Наполняем мок нативной плоской структуры ReaderStats библиотеки kafka-go
	expectedLibraryStats := kafka.ReaderStats{
		Topic:         "tiny.auth.login.attempts",
		Partition:     "0,1,2,3",
		Messages:      1500, // В kafka-go это int64
		Errors:        2,    // В kafka-go это int64
		Offset:        1050,
		Lag:           15,
		QueueLength:   5,
		QueueCapacity: 100,
	}
	mockLink.On("Stats").Return(expectedLibraryStats)

	receiver := &Receiver{
		opts: &ReceiverOptions{
			TargetName: "tiny.auth.login.attempts",
		},
		reader: mockLink,
	}

	// Act
	receiverStats := receiver.Stats()

	// Assert
	// 1. Проверяем блок COMMON Specific
	assert.Equal(t, "kafka", receiverStats.BrokerType)
	assert.Equal(t, "tiny.auth.login.attempts", receiverStats.TargetName)
	assert.Equal(t, "connected", receiverStats.Status)
	assert.Equal(t, uint64(1500), receiverStats.TotalMessages) // Проверяем честный uint64
	assert.Equal(t, uint64(2), receiverStats.TotalErrors)      // Проверяем честный uint64
	assert.Equal(t, int64(15), receiverStats.Lag)

	// 2. Проверяем блок KAFKA Specific
	assert.Equal(t, "0,1,2,3", receiverStats.Partition)
	assert.Equal(t, int64(1050), receiverStats.Offset)
	assert.Equal(t, int64(5), receiverStats.QueueLength)
	assert.Equal(t, int64(100), receiverStats.QueueCapacity)

	// 3. Проверяем, что блоки AMQP Specific остались нулевыми (скроются в JSON через omitempty)
	assert.Zero(t, receiverStats.ConsumerCount)
	assert.Zero(t, receiverStats.PrefetchCount)

	mockLink.AssertExpectations(t)
}
