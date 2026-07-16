package azure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	mocks3 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/azure/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSender_Publish_Success_Via_Connector(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// 1. Страхуем мок логгера от неожиданных вызовов Debug/Debugf
	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()

	mockConnector := mocks2.NewMockConnector[*amqp.Session](t)
	// ВАЖНО: Так как s.sender уже прогрет (Fast Path), метод getSender()
	// вообще не пойдет к коннектору. GetConnection НЕ должен вызываться!

	mockSenderLink := mocks3.NewMockAmqpSenderLink(t)
	mockSenderLink.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	opts := NewSenderOptions()
	opts.Logger = mockLogger
	opts.TargetName = "audit-topic"
	opts.Connector = mockConnector

	s, err := NewSender(func(cso *SenderOptions) {
		*cso = *opts
	})
	require.NoError(t, err)

	// Имитируем Fast Path — забиваем живой линк напрямую
	s.sender = mockSenderLink

	msg := &pkgamqp.Message[*amqp.MessageHeader]{Payload: []byte(`{"action":"login"}`)}

	// Act
	err = s.Publish(ctx, msg, nil)

	// Assert
	assert.NoError(t, err)
	mockSenderLink.AssertExpectations(t)
	mockConnector.AssertExpectations(t) // Убеждаемся, что GetConnection действительно не дергался
}

func TestSender_Publish_Retry_And_Invalidate(t *testing.T) {
	// Arrange
	ctx := context.Background()
	targetName := "audit-topic"

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockConnector := mocks2.NewMockConnector[*amqp.Session](t)
	linkErr := &amqp.LinkError{RemoteErr: &amqp.Error{Condition: amqp.ErrCondInternalError}}

	// 1. На первой попытке (attempt=1) handleSendError ОБЯЗАН вызвать Invalidate!
	mockConnector.On("Invalidate", linkErr).Return().Once()

	// 2. На второй попытке (attempt=2), так как локальный линк сброшен,
	// код пойдет в getSender() -> GetConnection(). Заставим его вернуть ошибку сессии,
	// чтобы Publish завершил цикл ретраев без всяких time.Sleep!
	expectedSessErr := errors.New("connector session lost permanently")
	mockConnector.On("GetConnection", mock.Anything).Return(nil, expectedSessErr).Once()

	mockSenderLink := mocks3.NewMockAmqpSenderLink(t)
	mockSenderLink.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(linkErr).Once()

	opts := NewSenderOptions()
	opts.Logger = mockLogger
	opts.TargetName = targetName
	opts.Connector = mockConnector
	opts.PublishMaxTryAttempts = 2 // 2 попытки
	opts.PublishBaseRetryDelay = 1 * time.Millisecond
	opts.PublishMaxRetryDelay = 2 * time.Millisecond

	s, err := NewSender(func(cso *SenderOptions) {
		*cso = *opts
	})
	require.NoError(t, err)

	// Прогреваем локальный линк для Fast Path на первой попытке
	s.sender = mockSenderLink

	// Act
	err = s.Publish(ctx, &pkgamqp.Message[*amqp.MessageHeader]{Payload: []byte(`{}`)}, nil)

	// Assert
	// Тест завершится с ошибкой инициализации соединения на 2-й попытке
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "azure sender failed to initialize link")

	// Проверяем, что локальный линк был успешно стерт на промежуточном этапе
	assert.Nil(t, s.sender, "Локальный линк должен быть обнулен при сетевом сбое")

	mockConnector.AssertExpectations(t)
	mockSenderLink.AssertExpectations(t)
}
