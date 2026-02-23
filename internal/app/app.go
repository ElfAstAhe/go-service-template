package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"github.com/hellofresh/health-go/v5"
)

type App struct {
	// app context
	ctx    context.Context
	cancel context.CancelFunc
	// app config
	config *config.Config
	// logging
	logger logger.Logger

	// DB
	db db.DB

	// infra
	wg sync.WaitGroup

	// checkers
	health *health.Health
	// http
	httpRouter transport.HTTPRouter
	httpServer *http.Server
}

func NewApp(config *config.Config, logger logger.Logger) *App {
	appCtx, appCancel := context.WithCancel(context.Background())

	return &App{
		ctx:    appCtx,
		cancel: appCancel,
		config: config,
		logger: logger,
	}
}

// Init инициализирует тяжелые ресурсы: БД, Кеш, Репозитории
func (app *App) Init() error {
	log := app.logger.GetLogger("App.Init")

	log.Info("initializing application resources...")

	log.Info("init helpers")
	if err := app.initHelpers(); err != nil {
		return err
	}

	log.Info("init database")
	if err := app.initDB(); err != nil {
		return err
	}

	log.Info("launch migrations")
	if err := app.migrateDB(); err != nil {
		return err
	}

	log.Info("init dependencies")
	if err := app.initDependencies(); err != nil {
		return err
	}

	log.Info("init startup services")
	if err := app.initStartupServices(); err != nil {
		return err
	}

	log.Info("init health")
	if err := app.initHealth(); err != nil {
		return err
	}

	log.Info("init http router")
	if err := app.initHTTPRouter(); err != nil {
		return err
	}

	log.Info("init http server")
	if err := app.initHTTPServer(); err != nil {
		return err
	}

	log.Info("init gRPC server")
	if err := app.initGRPCServer(); err != nil {
		return err
	}

	return nil
}

// Run запускает серверы (HTTP/gRPC) и блокирует поток до сигнала завершения
func (app *App) Run() error {
	log := app.logger.GetLogger("App.Run")
	log.Info("application is running")

	// Тут будет запуск http.ListenAndServe в горутине
	// и ожидание <-ctx.Done() или сигналов ОС

	return nil
}

// Stop - метод остановки приложения
func (app *App) Stop() {
	app.cancel()
}

func (app *App) WaitForStop() {
	app.wg.Wait()
}

// Close - метод освобождения ресурсов приложения
//
//	if err := app.Close(); err != nil {
//		log.Errorf("app close error [%v]", err)
//
//		panic(errs.NewAppCommonError("app close failed", err))
//	}
func (app *App) Close() error {
	log := app.logger.GetLogger("bootstrap close")

	//log.Info("save in mem data")
	//if err := app.saveInMemData(); err != nil {
	//    return err
	//}

	log.Info("close db connection")
	if err := app.db.Close(); err != nil {
		return err
	}

	//log.Info("close audit event service")
	//if err := app.auditEventService.Close(); err != nil {
	//    return err
	//}

	return nil
}

// gracefulShutdown - внутренний метод "агрессивного" закрытия приложения (ctrl+c) + остальные сигналы OS на закрытие
func (app *App) gracefulShutdown() {
	//    log := app.logger.GetLogger("App.gracefulShutdown")
	defer app.wg.Done()
	// channel
	sig := make(chan os.Signal, 1)
	// register channel signals
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// awaiting signal
	select {
	case <-sig:
		{
			app.cancel()
			break
		}
	case <-app.ctx.Done():
		{
			signal.Stop(sig)
			break
		}
	}

	// stop http
	//log.Info("graceful shutdown http server")
	//if err := app.httpServer.Shutdown(context.Background()); err != nil {
	//    log.Errorf("error graceful shutdown http server with error [%v]", err)
	//}
	// stop gRPC
	//log.Info("graceful shutdown gRPC server")
	//app.grpcServer.GracefulStop()
}
