package kafka

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/kafka/mocks"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSender_Publish_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaSenderLink(t)

	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger).Maybe()
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()

	transportMsg := &Message{
		TargetName: "tiny.auth.login.attempts",
		Payload:    []byte(`{"status":"success"}`),
		Props: map[string]any{
			"kafka_message_key": "user_abc",
			"custom_header_1":   "meta_value",
		},
	}

	// Настраиваем строгое ожидание вызова WriteMessages с проверкой среза сообщений
	mockLink.On("WriteMessages", ctx, mock.MatchedBy(func(msgs []kafka.Message) bool {
		// Проверяем, что передано ровно одно сообщение
		if len(msgs) != 1 {
			return false
		}

		msg := msgs[0]
		hasKey := string(msg.Key) == "user_abc"
		hasValue := string(msg.Value) == `{"status":"success"}`

		// Проверяем единственный заголовок в срезе заголовков
		hasHeader := len(msg.Headers) == 1 &&
			msg.Headers[0].Key == "custom_header_1" &&
			string(msg.Headers[0].Value) == "meta_value"

		return hasKey && hasValue && hasHeader
	})).Return(nil)

	sender := &Sender{
		opts: &SenderOptions{
			Brokers:               []string{"127.0.0.1:9092"},
			TargetName:            "tiny.auth.login.attempts",
			PublishMaxTryAttempts: 2,
		},
		writer: mockLink,
		logger: mockLogger,
	}

	// Act
	err := sender.Publish(ctx, transportMsg, nil)

	// Assert
	assert.NoError(t, err)
	mockLink.AssertExpectations(t)
}

func TestSender_Publish_RetryAndFallbackOnNetworkError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaSenderLink(t)

	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger).Maybe()
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return().Maybe()

	transportMsg := &Message{
		TargetName: "tiny.auth.login.attempts",
		Payload:    []byte(`{"status":"fail"}`),
	}

	// Создаем сетевую ошибку net.OpError, которая считается восстанавливаемой
	mockNetErr := &net.OpError{Op: "write", Net: "tcp", Err: errors.New("broken pipe")}

	// Имитируем падение по сети на ОБЕИХ попытках.
	// Так как у нас теперь нет "живого" врайтера, мы полностью контролируем его поведение через мок.
	mockLink.On("WriteMessages", ctx, mock.Anything).Return(mockNetErr).Times(2)

	sender := &Sender{
		opts: &SenderOptions{
			Brokers:               []string{"127.0.0.1:9092"},
			TargetName:            "tiny.auth.login.attempts",
			PublishMaxTryAttempts: 2, // 2 попытки
			PublishBaseRetryDelay: 1 * time.Millisecond,
			PublishMaxRetryDelay:  2 * time.Millisecond,
			ConnectTimeout:        5 * time.Second,
		},
		writer: mockLink,
		logger: mockLogger,
	}

	// Act
	err := sender.Publish(ctx, transportMsg, nil)

	// Assert
	// Ожидаем ошибку, так как все попытки исчерпаны
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka unrecoverable send error or retries exhausted")

	mockLink.AssertExpectations(t)
}
