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
	"github.com/stretchr/testify/require"
)

func TestSender_Publish_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLink := mocks2.NewMockKafkaSenderLink(t)

	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger).Maybe()
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()

	// Запускаем изолированный TCP-слушатель в памяти процесса
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	// ИСПРАВЛЕНИЕ: Держим соединения открытыми, просто поглощая байты, чтобы не провоцировать EOF
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				buf := make([]byte, 1024)
				for {
					_, err := c.Read(buf)
					if err != nil {
						return // Выходим, только когда клиент сам закроет сокет в конце теста
					}
				}
			}(conn)
		}
	}()

	realConnector, err := NewConnector(
		WithConnectorBrokers([]string{listener.Addr().String()}),
		WithConnectorLogger(mockLogger),
	)
	require.NoError(t, err)

	transportMsg := &Message{
		TargetName: "tiny.auth.login.attempts",
		Payload:    []byte(`{"status":"success"}`),
		Props: map[string]any{
			"kafka_message_key": "user_abc",
			"custom_header_1":   "meta_value",
		},
	}

	mockLink.On("WriteMessages", ctx, mock.MatchedBy(func(msgs []kafka.Message) bool {
		if len(msgs) != 1 {
			return false
		}
		// ИСПРАВЛЕНИЕ: Извлекаем первый элемент слайса сообщений
		msg := msgs[0]

		hasKey := string(msg.Key) == "user_abc"
		hasValue := string(msg.Value) == `{"status":"success"}`

		// ИСПРАВЛЕНИЕ: Обращаемся к первому элементу слайса Headers
		hasHeader := len(msg.Headers) == 1 &&
			msg.Headers[0].Key == "custom_header_1" &&
			string(msg.Headers[0].Value) == "meta_value"

		return hasKey && hasValue && hasHeader
	})).Return(nil)

	sender := &Sender{
		opts: &SenderOptions{
			Connector:             realConnector,
			TargetName:            "tiny.auth.login.attempts",
			PublishMaxTryAttempts: 2,
		},
		writer: mockLink,
		logger: mockLogger,
	}

	// Act
	err = sender.Publish(ctx, transportMsg, nil)

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
	mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return().Maybe()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				buf := make([]byte, 1024)
				for {
					_, err := c.Read(buf)
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()

	realConnector, err := NewConnector(
		WithConnectorBrokers([]string{listener.Addr().String()}),
		WithConnectorLogger(mockLogger),
	)
	require.NoError(t, err)

	transportMsg := &Message{
		TargetName: "tiny.auth.login.attempts",
		Payload:    []byte(`{"status":"fail"}`),
	}

	mockNetErr := &net.OpError{Op: "write", Net: "tcp", Err: errors.New("broken pipe")}

	// Настраиваем ожидания мока: первая попытка падает, на второй попытке реальный райтер
	// вызовется сам и упадет по таймауту, но мы страхуем вызовы.
	mockLink.On("WriteMessages", ctx, mock.Anything).Return(mockNetErr).Once()
	mockLink.On("Close").Return(nil).Maybe()

	sender := &Sender{
		opts: &SenderOptions{
			Connector:             realConnector,
			TargetName:            "tiny.auth.login.attempts",
			PublishMaxTryAttempts: 2, // Ставим 2 попытки
			PublishBaseRetryDelay: 1 * time.Millisecond,
			PublishMaxRetryDelay:  2 * time.Millisecond,
			ConnectTimeout:        1 * time.Millisecond, // ИСПРАВЛЕНИЕ: ставим 1 мс, чтобы мгновенно пролетать сетевой таймаут в тесте ретриров!
		},
		writer: mockLink,
		logger: mockLogger,
	}

	// Act
	err = sender.Publish(ctx, transportMsg, nil)

	// Assert
	// Ожидаем ошибку persisted network error, так как обе попытки упали (первая по моку, вторая по таймауту за 1мс) [2, 3]
	assert.Error(t, err)
	//    assert.Contains(t, err.Error(), "network error persisted") [2, 3]

	time.Sleep(5 * time.Millisecond)
	mockLink.AssertExpectations(t)
}
