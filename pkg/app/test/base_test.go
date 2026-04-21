package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/app"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/container/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBaseApplication_FullRun(t *testing.T) {
	mockLog := mocks.NewMockLogger(t)
	mockOrch := mocks2.NewMockOrchestrator(t)
	mockRunner := mocks2.NewMockRunner(t)

	// Настраиваем логи и оркестратор
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Infof", mock.Anything, mock.Anything).Return()
	mockLog.On("Info", mock.Anything).Return()
	mockLog.On("Debug", mock.Anything).Return()
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return()

	app := app.NewBaseApplication(
		app.WithOrchestrator(mockOrch),
		app.WithLogger(mockLog),
	)

	ctx := app.GetContext()

	// Настраиваем Init
	mockOrch.On("Init", mock.Anything).Return(nil).Once()

	// Настраиваем запуск раннеров
	mockOrch.On("GetRunners").Return([]container.Runner{mockRunner}, nil).Twice() // Для Start и Stop
	mockRunner.On("GetName").Return("test-runner")

	// Runner.Start должен блокироваться до отмены контекста
	mockRunner.On("Start", mock.MatchedBy(func(c context.Context) bool { return true })).
		Return(nil).
		Run(func(args mock.Arguments) {
			<-ctx.Done() // Ждем сигнала отмены извне
		}).Once()

	// Настраиваем Stop
	mockRunner.On("Stop", mock.Anything).Return(nil).Once()

	// 1. Init
	require.NoError(t, app.Init())

	// 2. Запускаем Run в горутине, так как он заблокируется на WaitForStop
	runErrChan := make(chan error, 1)
	go func() {
		runErrChan <- app.Run()
	}()

	// 3. Эмулируем приход сигнала (просто отменяем контекст вручную)
	// В реальном приложении это сделает GracefulShutdown, получив SIGTERM
	app.GetCancel()()

	// 4. Ждем завершения Run
	select {
	case err := <-runErrChan:
		assert.NoError(t, err)
	case <-time.After(time.Second * 2):
		t.Fatal("Application didn't stop in time")
	}

	mockRunner.AssertExpectations(t)
	mockOrch.AssertExpectations(t)
}

func TestBaseApplication_InitError(t *testing.T) {
	mockOrch := mocks2.NewMockOrchestrator(t)
	app := app.NewBaseApplication(app.WithOrchestrator(mockOrch))

	expectedErr := errors.New("db connection failed")
	mockOrch.On("Init", mock.Anything).Return(expectedErr).Once()

	err := app.Init()
	assert.ErrorIs(t, err, expectedErr)
}

func TestBaseApplication_RunnerFailure(t *testing.T) {
	// 1. Создаем моки
	mockLog := mocks.NewMockLogger(t)
	mockOrch := mocks2.NewMockOrchestrator(t)
	mockRunner1 := mocks2.NewMockRunner(t)
	mockRunner2 := mocks2.NewMockRunner(t)

	// 2. Важно! Настраиваем логгер, чтобы он не возвращал nil на GetLogger
	mockLog.On("GetLogger", mock.Anything).Return(mockLog)
	mockLog.On("Infof", mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLog.On("Info", mock.Anything).Return().Maybe()
	mockLog.On("Debug", mock.Anything).Return().Maybe()
	mockLog.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()

	// 3. Создаем приложение со всеми зависимостями
	app := app.NewBaseApplication(
		app.WithOrchestrator(mockOrch),
		app.WithLogger(mockLog), // ПЕРЕДАЕМ ЛОГГЕР, чтобы не было паники
	)

	// Настраиваем поведение оркестратора
	mockOrch.On("GetRunners").Return([]container.Runner{mockRunner1, mockRunner2}, nil)

	mockRunner1.On("GetName").Return("failing-runner")
	mockRunner2.On("GetName").Return("stable-runner")

	// Первый падает
	mockRunner1.On("Start", mock.Anything).Return(errors.New("crash")).Once()
	// Второй висит до отмены
	mockRunner2.On("Start", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		<-app.GetContext().Done()
	}).Once()

	// Настраиваем Stop для обоих (т.к. Run вызовет Stop в конце)
	mockRunner1.On("Stop", mock.Anything).Return(nil).Maybe()
	mockRunner2.On("Stop", mock.Anything).Return(nil).Maybe()

	// Запускаем
	go func() {
		_ = app.Run()
	}()

	// Проверяем результат
	select {
	case <-app.GetContext().Done():
		// OK: контекст отменился из-за падения runner1
	case <-time.After(app.GetConfig().StopTimeout):
		t.Fatal("Application should have canceled context")
	}
}
