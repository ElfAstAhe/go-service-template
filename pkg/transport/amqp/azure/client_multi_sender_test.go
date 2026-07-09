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

func TestClientMultiSender_Publish_Success_And_Cache(t *testing.T) {
	// Arrange
	ctx := context.Background()
	targetName := "audit-log-topic"

	// 1. Настраиваем мок логгера
	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", "azure-amqp-client-multi-sender").Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	// 2. Настраиваем мок AMQP линка
	mockSender := new(mocks2.MockAmqpSenderLink) // Сгенерированный mockery мок для amqpSenderLink
	// Ожидаем, что метод Send вызовется ровно 2 раза (для двух разных сообщений в один топик)
	mockSender.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	opts := NewClientSenderOptions()
	opts.Logger = mockLogger

	cms, err := NewClientMultiSender(func(cso *ClientSenderOptions) {
		*cso = *opts
	})
	require.NoError(t, err)

	// Вручную прогреваем кэш для конкретного топика, чтобы не дергать Dial
	cms.senders[targetName] = mockSender

	msg1 := &pkgamqp.Message[*amqp.MessageHeader]{Payload: []byte(`{"event":"auth_success"}`)}
	msg2 := &pkgamqp.Message[*amqp.MessageHeader]{Payload: []byte(`{"event":"auth_failed"}`)}

	// Act
	// Отправляем первое сообщение — должно пойти по Fast Path из кэша
	err = cms.Publish(ctx, targetName, nil, msg1, nil)
	assert.NoError(t, err)

	// Отправляем второе сообщение в тот же топик — проверяем, что кэш не инвалидировался
	err = cms.Publish(ctx, targetName, nil, msg2, nil)
	assert.NoError(t, err)

	// Assert
	mockSender.AssertExpectations(t)
	// Проверяем, что в списке активных топиков зарегистрирован наш топик
	assert.ElementsMatch(t, []string{targetName}, cms.GetTargetNames())
}

func TestClientMultiSender_Publish_RetryOnLinkError(t *testing.T) {
	// Arrange
	// Создаем контекст, который мы отменим, чтобы не уйти на реальный Dial на 2-й попытке
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	targetName := "metrics-queue"

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockSender := new(mocks2.MockAmqpSenderLink)

	// Симулируем сетевую ошибку линка AMQP на первой попытке
	linkErr := &amqp.LinkError{RemoteErr: &amqp.Error{Condition: amqp.ErrCondInternalError}}
	mockSender.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(linkErr).Once()

	opts := NewClientSenderOptions()
	opts.Logger = mockLogger
	opts.TargetName = "default-fallback-queue"
	opts.PublishMaxTryAttempts = 2 // Ставим 2 попытки, чтобы 1-я считалась временным сбоем
	opts.PublishBaseRetryDelay = 1 * time.Millisecond
	opts.PublishMaxRetryDelay = 2 * time.Millisecond

	cms, err := NewClientMultiSender(func(cso *ClientSenderOptions) {
		*cso = *opts
	})
	require.NoError(t, err)

	// Забиваем мок сендера в кэш
	cms.senders[targetName] = mockSender

	msg := &pkgamqp.Message[*amqp.MessageHeader]{Payload: []byte(`{"cpu": 42}`)}

	// Запускаем фоновую горутину, которая отменит контекст через мгновение,
	// как только handleSendError отработает и очистит карту, но ДО того, как начнется 2-я попытка.
	go func() {
		time.Sleep(2 * time.Millisecond)
		cancel()
	}()

	// Act
	err = cms.Publish(ctx, targetName, nil, msg, nil)

	// Assert
	// Тест должен завершиться с ошибкой, так как мы отменили контекст
	assert.Error(t, err)

	// САМАЯ ВАЖНАЯ ПРОВЕРКА: Поскольку 1 < 2, handleSendError ОБЯЗАН был
	// зайти в ветку switch и удалить битый топик из карты сендеров!
	cms.mu.RLock()
	_, exists := cms.senders[targetName]
	cms.mu.RUnlock()
	assert.False(t, exists, "Битый линк должен быть успешно удален из кэша сендеров на промежуточной попытке")

	mockSender.AssertExpectations(t)
}

func TestClientMultiSender_Close_ParallelLinksClosing(t *testing.T) {
	// Arrange
	ctx := context.Background()

	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything).Return().Maybe()

	// Создаем два мока линков для симуляции отправки в разные топики
	mockSender1 := new(mocks2.MockAmqpSenderLink)
	mockSender2 := new(mocks2.MockAmqpSenderLink)

	// Настраиваем Close для обоих линков с искусственной задержкой
	mockSender1.On("Close", mock.Anything).Return(nil).After(10 * time.Millisecond)
	mockSender2.On("Close", mock.Anything).Return(nil).After(10 * time.Millisecond)

	opts := NewClientSenderOptions()
	opts.Logger = mockLogger
	opts.ShutdownTimeout = 50 * time.Millisecond

	cms, err := NewClientMultiSender(func(cso *ClientSenderOptions) {
		*cso = *opts
	})
	require.NoError(t, err)

	// Забиваем пул двумя «прогретыми» топиками
	cms.senders["topic-a"] = mockSender1
	cms.senders["topic-b"] = mockSender2

	start := time.Now()

	// Act
	err = cms.Close(ctx)

	duration := time.Since(start)

	// Assert
	assert.NoError(t, err)
	// Проверяем конкурентность: если бы линки закрывались последовательно,
	// метод Close() занял бы не менее 20ms (10ms + 10ms).
	// Благодаря sync.WaitGroup они закроются параллельно за ~10ms.
	assert.Less(t, duration, 18*time.Millisecond, "Линки закрываются последовательно, а не параллельно!")

	// Карта должна атомарно очиститься
	assert.Empty(t, cms.GetTargetNames())

	mockSender1.AssertExpectations(t)
	mockSender2.AssertExpectations(t)
}
