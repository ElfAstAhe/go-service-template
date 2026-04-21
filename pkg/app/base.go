package app

import (
	"context"
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
		return errs.NewCommonError("app orchestrator is nil", nil)
	}
	if err := app.orchestrator.Init(app.ctx); err != nil {
		return err
	}

	return nil
}

func (app *BaseApplication) Run() error {
	// ToDo: implement

	return nil
}

func (app *BaseApplication) Start() error {
	// ToDo: implement

	return nil
}

func (app *BaseApplication) Stop() error {
	app.cancel()

	return nil
}

func (app *BaseApplication) Close() error {
	// ToDo: implement

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
