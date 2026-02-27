package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "expvar"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/domain"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
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

	// tx
	tm db.TransactionManager

	// repositories
	// test repo
	testRepo domain.TestRepository

	// use cases
	// test get
	testGetUC usecase.TestGetUseCase
	// test get by code
	testGetByCodeUC usecase.TestGetByCodeUseCase
	// test list
	testListUC usecase.TestListUseCase
	// test save
	testSaveUC usecase.TestSaveUseCase
	// test delete
	testDeleteUC usecase.TestDeleteUseCase

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
	log := app.logger.GetLogger("App.Close")

	log.Info("close test repository")
	if err := app.testRepo.Close(); err != nil {
		return err
	}

	log.Info("close db connection")
	if err := app.db.Close(); err != nil {
		return err
	}

	return nil
}

// gracefulShutdown - внутренний метод "агрессивного" закрытия приложения (ctrl+c) + остальные сигналы OS на закрытие
func (app *App) gracefulShutdown() {
	defer app.wg.Done()

	log := app.logger.GetLogger("App.gracefulShutdown")
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

	var shutdownWg sync.WaitGroup

	shutdownWg.Add(1)
	go func() { // stop HTTP
		defer shutdownWg.Done()

		ctxTimed, cancelTimed := context.WithTimeout(context.Background(), app.config.HTTP.ShutdownTimeout)
		defer cancelTimed()

		// stop http
		log.Info("shutdown http server...")
		if err := app.httpServer.Shutdown(ctxTimed); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Error("http server shutdown timed out (force close)")
			} else {
				log.Warnf("shutdown http server with error [%v]", err)
			}
		}
		log.Info("shutdown http server complete")
	}()

	shutdownWg.Add(1)
	go func() { // stop gRPC
		defer shutdownWg.Done()

		log.Info("shutdown gRPC server...")

		doneChan := make(chan struct{})
		go func() {
			app.grpcServer.GracefulStop()
			close(doneChan)
		}()
		select {
		case <-doneChan:
			log.Info("shutdown gRPC server complete")
		case <-time.After(app.config.HTTP.ShutdownTimeout):
			log.Error("gRPC graceful shutdown timed out: forcing stop")
			app.grpcServer.Stop()
		}
	}()

	// Ожидаем завершения остановки всех серверов
	shutdownWg.Wait()
}
