package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "expvar"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	_ "github.com/ElfAstAhe/go-service-template/migrations/example-service"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport"
	"github.com/hellofresh/health-go/v5"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
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
	// gRPC
	grpcServer *grpc.Server
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

	log.Info("start graceful shutdown")
	app.wg.Add(1)
	go app.gracefulShutdown()

	var eg errgroup.Group
	log.Info("start servers...")
	// http
	eg.Go(func() error {
		if err := app.launchHTTPServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})
	//// gRPC
	//eg.Go(func() error {
	//    if err := app.launchGRPCServer(); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
	//        log.Errorf("Error starting gRPC server with error [%v]", err)
	//
	//        return err
	//    }
	//
	//    return nil
	//})
	//
	return eg.Wait()
}

func (app *App) launchHTTPServer() error {
	log := app.logger.GetLogger("App.launchHTTPServer")
	if app.config.HTTP.Secure {
		log.Info("enable https")

		return app.httpServer.ListenAndServeTLS(app.config.HTTP.CertificatePath, app.config.HTTP.PrivateKeyPath)
	}

	log.Info("enable http")

	return app.httpServer.ListenAndServe()
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

	log.Info("close db connection")
	if err := app.db.Close(); err != nil {
		return err
	}

	return nil
}

// gracefulShutdown - внутренний метод "агрессивного" закрытия приложения (ctrl+c) + остальные сигналы OS на закрытие
func (app *App) gracefulShutdown() {
	log := app.logger.GetLogger("App.gracefulShutdown")
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

	ctxTimed, _ := context.WithTimeout(context.Background(), app.config.HTTP.ShutdownTimeout)

	// stop http
	log.Info("shutdown http server...")
	if err := app.httpServer.Shutdown(ctxTimed); err != nil {
		log.Warnf("shutdown http server with error [%v]", err)
	}
	log.Info("shutdown http server complete")

	// stop gRPC
	log.Info("shutdown gRPC server...")
	//app.grpcServer.GracefulStop()
	log.Info("shutdown gRPC server complete")
}
