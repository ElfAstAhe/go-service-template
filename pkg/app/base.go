package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type BaseApplication struct {
	// config
	conf *config.AppConfig
	// orchestrator containers orchestrator
	orchestrator container.Orchestrator
	// logging
	logger logger.Logger
	// application ctx with cancel
	ctx    context.Context
	cancel context.CancelFunc
	// wg
	wg sync.WaitGroup
}

func NewBaseApplication(opts ...Option) *BaseApplication {
	// app level context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	// new app instance with defaults
	res := &BaseApplication{
		conf:   config.NewDefaultAppConfig(),
		ctx:    ctx,
		cancel: cancel,
	}
	// setup instance
	for _, opt := range opts {
		opt(res)
	}

	return res
}

func (app *BaseApplication) Init() error {
	// orchestrator
	if utils.IsNil(app.orchestrator) {
		return errs.NewCommonError("orchestrator is nil", nil)
	}

	initCtx, cancel := context.WithTimeout(app.ctx, app.conf.InitTimeout)
	defer cancel()

	if err := app.orchestrator.Init(initCtx); err != nil {
		return errs.NewCommonError("orchestrator init failed", err)
	}

	return nil
}

func (app *BaseApplication) Run() error {
	// 1. Запускаем Runners (переводим их в состояние Running)
	if err := app.Start(); err != nil {
		return errs.NewCommonError("failed to start runners", err)
	}

	// 2. Включаем слушатель сигналов ОС в отдельной горутине
	app.wg.Add(1)
	go app.GracefulShutdown()

	// 3. Ожидаем отмены контекста
	app.logger.Info("application is running and waiting for app context cancel")
	<-app.ctx.Done()

	// 4. Останавливаем runners
	app.logger.Info("application is shutting down")
	err := app.Stop()

	// Ожидаем остановки всех горутин
	app.logger.Info("application is waiting for runners stopped")
	app.WaitForStop()

	return err
}

func (app *BaseApplication) Start() error {
	runners, err := app.orchestrator.GetRunners()
	if err != nil {
		return err
	}

	for _, r := range runners {
		app.wg.Add(1)
		go func(runner container.Runner) {
			defer app.wg.Done()
			app.logger.Infof("runner [%s] starting", runner.GetName())

			if err := runner.Start(app.ctx); err != nil {
				app.logger.Errorf("runner [%s] failed: %v", runner.GetName(), err)
				app.cancel() // Даем команду на выход всему приложению
			}
		}(r)
	}

	return nil
}

func (app *BaseApplication) Stop() error {
	app.logger.Info("stopping active runners (graceful shutdown phase)...")

	// На всякий случай дублируем отмену контекста
	app.cancel()

	runners, err := app.orchestrator.GetRunners()
	if err != nil {
		return err
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), app.conf.StopTimeout)
	defer stopCancel()

	var (
		stopWg   sync.WaitGroup
		mu       sync.Mutex
		stopErrs []error
	)

	for _, r := range runners {
		stopWg.Add(1)
		go func(runner container.Runner) {
			defer stopWg.Done()
			// Каждый Runner знает свой stop timeout
			if err := runner.Stop(stopCtx); err != nil {
				mu.Lock()
				stopErrs = append(stopErrs, err)
				mu.Unlock()
			}
		}(r)
	}

	stopWg.Wait()
	return errors.Join(stopErrs...)
}

func (app *BaseApplication) Close() error {
	app.logger.Info("closing application resources (containers)...")

	// 1. Создаем контекст с таймаутом специально для фазы закрытия
	// Используем конфиг, который мы прокинули в BaseApplication
	closeCtx, cancel := context.WithTimeout(context.Background(), app.conf.CloseTimeout)
	defer cancel()

	// 2. Делегируем всё оркестратору
	// Он пройдёт по всем контейнерам в порядке LIFO
	if err := app.orchestrator.Close(closeCtx); err != nil {
		return errs.NewCommonError("orchestrator close failed", err)
	}

	app.logger.Info("application resources closed successfully")
	return nil
}

// GracefulShutdown метод мягкого закрытия приложения,
// слушает сигналы ос или контекст приложения, в случае сигнала OS отменяет контекст приложения,
// по приходу сигнала или отмены контеста приложения метод завершает свою работу
//
// Пример использования:
//
//	go app.GracefullShutdown()
func (app *BaseApplication) GracefulShutdown() {
	defer app.wg.Done()

	log := app.logger.GetLogger("BaseApplication.GracefulShutdown")

	log.Debugf("Graceful shutdown goroutine started")
	// channel
	osSigChan := make(chan os.Signal, 1)
	// signals
	osSignals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}
	// register channel signals
	signal.Notify(osSigChan, osSignals...)
	defer signal.Stop(osSigChan)
	//
	log.Debug("listening signals")
	for _, sig := range osSignals {
		log.Debugf("os signal: [%s]", sig.String())
	}
	// awaiting signal
	select {
	case osSig := <-osSigChan:
		log.Debugf("received os signal [%s], cancel main app context", osSig.String())
		app.cancel()
	case <-app.ctx.Done():
		log.Debug("main app context has been canceled")
	}
}

func (app *BaseApplication) WaitForStop() {
	app.wg.Wait()
}

func (app *BaseApplication) GetWaitGroup() *sync.WaitGroup {
	return &app.wg
}

func (app *BaseApplication) GetLogger() logger.Logger {
	return app.logger
}

func (app *BaseApplication) GetOrchestrator() container.Orchestrator {
	return app.orchestrator
}

func (app *BaseApplication) GetContext() context.Context {
	return app.ctx
}

func (app *BaseApplication) GetCancel() context.CancelFunc {
	return app.cancel
}

func (app *BaseApplication) GetConfig() *config.AppConfig {
	return app.conf
}
