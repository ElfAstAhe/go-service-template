package azure

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Безопасный стаб для интерфейса amqpReceiverLink, который теперь умеет всё
type fakeReceiverLink struct{}

func (f *fakeReceiverLink) Receive(ctx context.Context, opts *amqp.ReceiveOptions) (*amqp.Message, error) {
	return &amqp.Message{}, nil
}

func (f *fakeReceiverLink) AcceptMessage(ctx context.Context, msg *amqp.Message) error {
	return nil // Безопасный стаб
}

func (f *fakeReceiverLink) RejectMessage(ctx context.Context, msg *amqp.Message, err *amqp.Error) error {
	return nil // Безопасный стаб
}

func (f *fakeReceiverLink) ReleaseMessage(ctx context.Context, msg *amqp.Message) error {
	return nil // Безопасный стаб
}

func (f *fakeReceiverLink) Close(ctx context.Context) error {
	return nil // Гарантированная защита от паники
}

func createFakeSysMessage() *amqp.Message {
	return &amqp.Message{
		Data: [][]byte{
			[]byte(`{"user`),
			[]byte(`_id":`),
			[]byte(`"123"}`),
		},
		ApplicationProperties: map[string]any{
			"trace_id": "fake-trace-888",
		},
	}
}

// 1. Тест применения опций
func TestClientReceiver_OptionsApplication(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	cr := NewClientReceiver(
		"amqp://localhost:5672",
		mockLog,
		WithShutdownTimeout(10*time.Second),
	)

	assert.NotNil(t, cr)
	assert.Equal(t, 10*time.Second, cr.opts.shutdownTimeout)
}

// 2. Тест успешного извлечения контекста сообщения
func TestClientReceiver_ExtractAndAccept(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	cr := NewClientReceiver("amqp://localhost:5672", mockLog)

	sysMsg := createFakeSysMessage()
	cleanMsg := &pkgamqp.Message{
		Payload: []byte(`{"user_id":"123"}`),
		Properties: map[string]any{
			sysMsgKey: sysMsg,
		},
	}

	extracted, err := cr.extractOriginalMessage(cleanMsg)
	assert.NoError(t, err)
	assert.Equal(t, sysMsg, extracted)
	assert.Len(t, extracted.Data, 3)
}

// 3. Тест валидации ошибок извлечения
func TestClientReceiver_ExtractSysMessage_ValidationErrors(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	cr := NewClientReceiver("amqp://localhost:5672", mockLog)

	_, err := cr.extractOriginalMessage(nil)
	assert.Error(t, err)

	_, err = cr.extractOriginalMessage(&pkgamqp.Message{Payload: []byte("{}")})
	assert.Error(t, err)

	invalidMsg := &pkgamqp.Message{
		Payload: []byte("{}"),
		Properties: map[string]any{
			sysMsgKey: "corrupted-string",
		},
	}
	_, err = cr.extractOriginalMessage(invalidMsg)
	assert.Error(t, err)
}

// 4. Тест: Каскадный сброс стейт-машины при смерти сокета (ConnError) БЕЗ ПАНИК
func TestClientReceiver_HandleReceiverFailure(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	// Ожидаем вызов Errorf с вариативными аргументами
	mockLog.On("Errorf",
		"AMQP Receiver Socket dead: %v. Resetting server connection pipelines.",
		mock.Anything,
	).Once().Return()

	cr := NewClientReceiver("amqp://localhost:5672", mockLog)

	cr.conn = &amqp.Conn{}
	cr.session = &amqp.Session{}

	// ИСПРАВЛЕНО: Забиваем мапу нашим безопасным интерфейсным фейк-линком
	cr.receivers["audit.queue"] = &fakeReceiverLink{}

	connErr := &amqp.ConnError{
		RemoteErr: &amqp.Error{Condition: "amqp:connection:forced-close"},
	}

	// Теперь этот вызов отработает идеально, безопасно удалив фейк-линк из мапы
	cr.handleReceiverFailure("audit.queue", connErr)

	assert.Empty(t, cr.receivers)
	assert.Nil(t, cr.session)
	assert.Nil(t, cr.conn)
}
