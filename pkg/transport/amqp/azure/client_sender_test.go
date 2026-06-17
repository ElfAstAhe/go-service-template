package azure

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/goleak"
)

// Проверяем утечки горутин драйвера
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// Простейший безопасный стаб для интерфейса линка
type fakeLink struct{}

func (f *fakeLink) Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error {
	return nil
}
func (f *fakeLink) Close(ctx context.Context) error {
	return nil // Безопасно, без паник рантайма!
}

// 1. Тест применения Functional Options
func TestClientSender_OptionsApplication(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	cs := NewClientSender(
		"amqp://localhost:5672",
		mockLog,
		WithShutdownTimeout(10*time.Second),
		WithMaxFrameSize(4096),
	)

	assert.NotNil(t, cs)
	assert.Equal(t, 10*time.Second, cs.opts.shutdownTimeout)
	assert.Equal(t, uint32(4096), cs.opts.ConnOptions.MaxFrameSize)
}

// 2. Тест: Анализ и гранулярный сброс ресурсов при ошибке линка (LinkError)
func TestClientSender_HandleSendError_LinkFailure(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	// В боевом коде передается: формат, строка адреса (1) и сама ошибка (2).
	// Поэтому в On() мы пишем строку формата и два mock.Anything под аргументы.
	mockLog.On("Errorf",
		"AMQP Link dead for address %s: %v. Cleaning target link.",
		mock.Anything,
		mock.Anything,
	).Once().Return()

	cs := NewClientSender("amqp://localhost:5672", mockLog)

	cs.conn = &amqp.Conn{}
	cs.session = &amqp.Session{}
	cs.senders["auth.events"] = &fakeLink{}
	cs.senders["other.events"] = &fakeLink{}

	linkErr := &amqp.LinkError{RemoteErr: &amqp.Error{Condition: "amqp:link:detach-forced"}}
	cs.handleSendError("auth.events", linkErr)

	assert.Nil(t, cs.senders["auth.events"])
	assert.NotNil(t, cs.senders["other.events"])
	assert.NotNil(t, cs.session)
	assert.NotNil(t, cs.conn)
}

// 3. Тест: Каскадный сброс сессии и всех линков при SessionError
func TestClientSender_HandleSendError_SessionFailure(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	// В боевом коде передается: формат и ошибка (1 аргумент).
	mockLog.On("Errorf",
		"AMQP Session dead: %v. Invalidating session and all target links.",
		mock.Anything,
	).Once().Return()

	cs := NewClientSender("amqp://localhost:5672", mockLog)

	cs.conn = &amqp.Conn{}
	cs.session = &amqp.Session{}
	cs.senders["queue-1"] = &fakeLink{}
	cs.senders["queue-2"] = &fakeLink{}

	sessionErr := &amqp.SessionError{RemoteErr: &amqp.Error{Condition: "amqp:session:handle-in-use"}}
	cs.handleSendError("queue-1", sessionErr)

	assert.Empty(t, cs.senders)
	assert.Nil(t, cs.session)
	assert.NotNil(t, cs.conn)
}

// 4. Тест: Полное уничтожение всей стейт-машины при смерти сокета (ConnError)
func TestClientSender_HandleSendError_ConnectionFailure(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	// В боевом коде передается: формат и ошибка (1 аргумент).
	mockLog.On("Errorf",
		"AMQP Socket dead (or idle timeout): %v. Resetting entire cluster connection.",
		mock.Anything,
	).Once().Return()

	cs := NewClientSender("amqp://localhost:5672", mockLog)

	cs.conn = &amqp.Conn{}
	cs.session = &amqp.Session{}
	cs.senders["queue-a"] = &fakeLink{}

	connErr := &amqp.ConnError{
		RemoteErr: &amqp.Error{
			Condition:   "amqp:connection:forced-close",
			Description: "idle timeout expired",
		},
	}
	cs.handleSendError("queue-a", connErr)

	assert.Empty(t, cs.senders)
	assert.Nil(t, cs.session)
	assert.Nil(t, cs.conn)
}

// 5. Тест: Игнорирование обычных таймаутов контекста диспетчера
func TestClientSender_HandleSendError_IgnoreContextTimeout(t *testing.T) {
	mockLog := &mocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)

	cs := NewClientSender("amqp://localhost:5672", mockLog)

	cs.conn = &amqp.Conn{}
	cs.session = &amqp.Session{}
	cs.senders["stable-queue"] = &fakeLink{}

	cs.handleSendError("stable-queue", context.DeadlineExceeded)

	assert.NotNil(t, cs.senders["stable-queue"])
	assert.NotNil(t, cs.session)
	assert.NotNil(t, cs.conn)
}
