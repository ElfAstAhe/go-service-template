package azure

import (
	"context"
	"testing"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	mocks3 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/azure/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReceiver_Receive_Success_And_Payload(t *testing.T) {
	// Arrange
	ctx := context.Background()
	queueName := "audit-queue"

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)

	mockConnector := mocks2.NewMockConnector[*amqp.Session](t)

	// Симулируем пакет из двух чанков
	mockAzureMsg := &amqp.Message{
		Data: [][]byte{
			[]byte("part1_"),
			[]byte("part2"),
		},
	}

	mockReceiverLink := mocks3.NewMockAmqpReceiverLink(t)
	mockReceiverLink.On("Receive", mock.Anything, mock.Anything).Return(mockAzureMsg, nil).Once()

	opts := NewReceiverOptions()
	opts.Logger = mockLogger
	opts.TargetName = queueName
	opts.Connector = mockConnector

	r, err := NewReceiver(func(cro *ReceiverOptions) {
		*cro = *opts
	})
	require.NoError(t, err)
	r.link = mockReceiverLink

	// Act
	msg, err := r.Receive(ctx, nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []byte("part1_part2"), msg.Payload)
	assert.Equal(t, queueName, msg.TargetName) // Убеждаемся, что TargetName на месте
	mockReceiverLink.AssertExpectations(t)
}

func TestReceiver_Receive_Failure_InvalidatesConnector(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Warnf", mock.Anything, mock.Anything).Return().Maybe()

	mockConnector := mocks2.NewMockConnector[*amqp.Session](t)
	connErr := &amqp.ConnError{} // Ошибка сокета

	// Проверяем, что ресивер при ошибке чтения ЧЕСТНО вызвал Invalidate на общем коннекторе
	mockConnector.On("Invalidate", connErr).Return().Once()

	mockReceiverLink := mocks3.NewMockAmqpReceiverLink(t)
	mockReceiverLink.On("Receive", mock.Anything, mock.Anything).Return(nil, connErr).Once()

	opts := NewReceiverOptions()
	opts.Logger = mockLogger
	opts.TargetName = "audit-queue"
	opts.Connector = mockConnector

	r, err := NewReceiver(func(cro *ReceiverOptions) {
		*cro = *opts
	})
	require.NoError(t, err)
	r.link = mockReceiverLink

	// Act
	msg, err := r.Receive(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.Nil(t, r.link, "Локальный линк ресивера должен быть обнулен для ленивого реконнекта")
	mockConnector.AssertExpectations(t)
	mockReceiverLink.AssertExpectations(t)
}
