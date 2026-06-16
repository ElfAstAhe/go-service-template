package test

import (
	"context"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/pubsub"
	pubsubmocks "github.com/ElfAstAhe/go-service-template/pkg/infra/pubsub/mocks"
	loggermocks "github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/mock"
	"go.uber.org/goleak"
)

// Проверяем утечки горутин после выполнения всех тестов в пакете
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// 1. Тест успешной параллельной доставки события
func TestEventDispatcher_SuccessNotify(t *testing.T) {
	mockLog := &loggermocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	dispatcher := pubsub.NewEventDispatcher[string]("test-dispatcher", 50*time.Millisecond, mockLog)

	obs := &pubsubmocks.MockObserver[string]{}
	obs.On("GetName").Return("ArtemisObserver")

	done := make(chan struct{})

	// Ожидаем вызов OnNotify ровно один раз
	obs.On("OnNotify", mock.Anything, "payload-data").Return(nil).Run(func(args mock.Arguments) {
		close(done)
	})

	dispatcher.Register(obs)
	dispatcher.Notify(context.Background(), "payload-data")

	select {
	case <-done:
		// Успешно выполнено в фоне
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for observer execution")
	}

	obs.AssertExpectations(t)
}

// 2. Тест изоляции паники (Recover) и записи её в Errorf с новым форматом логов

func TestEventDispatcher_PanicRecovery(t *testing.T) {
	mockLog := &loggermocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	panicLogged := make(chan struct{})

	// Настраиваем логгер: используем mock.Anything для вариативного среза аргументов
	mockLog.On("Errorf",
		"pub/sub event dispatcher %s observer %s panic recovery %v",
		mock.Anything, // Забиваем на точный срез []any, так как testify падает на variadic параметрах
	).Run(func(args mock.Arguments) {
		close(panicLogged)
	}).Return() // Обязательно добавляем Return(), если метод mockery-мока что-то возвращает

	dispatcher := pubsub.NewEventDispatcher[string]("panic-dispatcher", 50*time.Millisecond, mockLog)

	obs := &pubsubmocks.MockObserver[string]{}
	obs.On("GetName").Return("FailingObserver")

	obs.On("OnNotify", mock.Anything, "bad-data").Run(func(args mock.Arguments) {
		panic("runtime memory failure simulation")
	})

	dispatcher.Register(obs)
	dispatcher.Notify(context.Background(), "bad-data")

	select {
	case <-panicLogged:
		// Паника изолирована и попала в Errorf, приложение стабильно
	case <-time.After(200 * time.Millisecond):
		t.Fatal("panic was not recovered or logged within timeframe")
	}

	mockLog.AssertExpectations(t)
}

// 3. Тест контроля таймаута фоновой группы
func TestEventDispatcher_TimeoutEnforcement(t *testing.T) {
	mockLog := &loggermocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	timeoutLogged := make(chan struct{})

	// Настраиваем логгер для теста таймаута
	mockLog.On("Errorf",
		"pub/sub event dispatcher %s observer %s on notify got error %v",
		mock.Anything,
	).Run(func(args mock.Arguments) {
		close(timeoutLogged)
	}).Return()

	dispatcher := pubsub.NewEventDispatcher[string]("timeout-dispatcher", 5*time.Millisecond, mockLog)

	obs := &pubsubmocks.MockObserver[string]{}
	obs.On("GetName").Return("SlowObserver")

	obs.On("OnNotify", mock.Anything, "slow-data").Return(func(ctx context.Context, data string) error {
		<-ctx.Done()
		return ctx.Err() // Возвращает context.DeadlineExceeded
	})

	dispatcher.Register(obs)
	dispatcher.Notify(context.Background(), "slow-data")

	select {
	case <-timeoutLogged:
		// Движок успешно прервал выполнение медленного подписчика через 5мс
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout was not enforced by the dispatcher")
	}
}

// 4. Тест динамического удаления подписчика (Unregister)
func TestEventDispatcher_Unregister(t *testing.T) {
	mockLog := &loggermocks.MockLogger{}
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	dispatcher := pubsub.NewEventDispatcher[string]("unreg-dispatcher", 50*time.Millisecond, mockLog)

	obs := &pubsubmocks.MockObserver[string]{}
	obs.On("GetName").Return("TemporaryObserver")

	dispatcher.Register(obs)
	dispatcher.Unregister(obs)

	dispatcher.Notify(context.Background(), "test-data")

	time.Sleep(20 * time.Millisecond)

	// Убеждаемся, что метод OnNotify вообще не вызывался
	obs.AssertNotCalled(t, "OnNotify", mock.Anything, mock.Anything)
}
