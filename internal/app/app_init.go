package app

import (
	"context"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/facade"
	"github.com/ElfAstAhe/go-service-template/internal/repository"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	grpcsvc "github.com/ElfAstAhe/go-service-template/internal/transport/grpc"
	"github.com/ElfAstAhe/go-service-template/internal/transport/rest"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	pb "github.com/ElfAstAhe/go-service-template/pkg/api/grpc/example/v1"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/telemetry"
	migrations "github.com/ElfAstAhe/go-service-template/pkg/migration/goose"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"github.com/hellofresh/health-go/v5"
	healthPg "github.com/hellofresh/health-go/v5/checks/pgx5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

func (app *App) initDB() error {
	var err error
	app.db, err = postgres.NewPgDB(app.config.DB)
	if err != nil {
		return errs.NewCommonError("init database error", err)
	}

	return nil
}

func (app *App) migrateDB() error {
	migrator, err := migrations.NewGooseDBMigrator(app.ctx, app.db, app.logger)
	if err != nil {
		return errs.NewCommonError("create migrator", err)
	}
	if err = migrator.Initialize(); err != nil {
		return errs.NewCommonError("init migrator", err)
	}
	if err = migrator.Up(); err != nil {
		return errs.NewCommonError("migrator up", err)
	}

	return nil
}

func (app *App) initTelemetry() error {
	// Вызываем нашу настройку
	shutdown, err := telemetry.SetupOTel(app.ctx, app.config.Telemetry)
	if err != nil {
		return errs.NewCommonError("failed to setup telemetry", err)
	}

	// Сохраняем shutdown в App, чтобы вызвать его в конце main
	app.telemetryShutdown = shutdown

	return nil
}

func (app *App) initMetrics() error {
	// Регистрация стандартных метрик Go (Memory, Goroutines, GC, Stack)
	// Они автоматически полетят в prometheus.DefaultRegisterer
	//if err := prometheus.Register(collectors.NewGoCollector()); err != nil {
	//    return errs.NewCommonError("failed to register go run-time collector", err)
	//}

	// Регистрация метрик процесса (CPU, Open FDs, Threads)
	//if err := prometheus.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
	//	return errs.NewCommonError("failed to register process collector", err)
	//}

	if app.db != nil {
		if err := prometheus.Register(collectors.NewDBStatsCollector(app.db.GetDB(), config.AppName)); err != nil {
			return errs.NewCommonError("failed to register db stats", err)
		}
	}

	return nil
}

func (app *App) initDependencies() error {
	// tx
	app.tm = db.NewTxManager(app.db)

	// helpers
	if err := app.initHealth(); err != nil {
		return errs.NewCommonError("init helpers", err)
	}

	// repositories
	if err := app.initRepositories(); err != nil {
		return errs.NewCommonError("init repositories", err)
	}

	// use cases
	if err := app.initUseCases(); err != nil {
		return errs.NewCommonError("init use cases", err)
	}

	// facades
	if err := app.initFacades(); err != nil {
		return errs.NewCommonError("init facades", err)
	}

	return nil
}

func (app *App) initHelpers() error {
	// here initialize any helpers
	// ..

	return nil
}

func (app *App) initRepositories() error {
	var err error
	// test repo
	app.testRepo, err = postgres.NewTestRepository(app.db, app.db)
	if err != nil {
		return errs.NewCommonError("create test repository", err)
	}
	// metrics test repo
	app.testRepo = repository.NewTestMetricsRepository(app.testRepo)

	return nil
}

func (app *App) initUseCases() error {
	// test get
	app.testGetUC = usecase.NewTestGetUseCase(app.testRepo)
	// test get by code
	app.testGetByCodeUC = usecase.NewTestGetByCodeUseCase(app.testRepo)
	// list
	app.testListUC = usecase.NewTestListUseCase(app.testRepo)
	// save
	app.testSaveUC = usecase.NewTestSaveUseCase(app.tm, app.testRepo)
	// test delete
	app.testDeleteUC = usecase.NewTestDeleteUseCase(app.tm, app.testRepo)

	return nil
}

func (app *App) initFacades() error {
	// test facade
	app.testFacade = facade.NewTestFacade(app.testGetUC, app.testGetByCodeUC, app.testListUC, app.testSaveUC, app.testDeleteUC)

	return nil
}

func (app *App) initStartupServices() error {
	// here initialize any startup services (workers, observers, etc.)
	// ..

	return nil
}

func (app *App) initHealth() error {
	healthChecker, err := health.New(health.WithComponent(health.Component{
		Name:    config.AppName,
		Version: config.AppVersion,
	}))
	if err != nil {
		return errs.NewCommonError("failed create health checker", err)
	}

	// Регистрируем Postgres. Либа сама будет делать Ping
	err = healthChecker.Register(health.Config{
		Name:      app.db.GetDriver(),
		Timeout:   app.config.DB.ConnTimeout,
		SkipOnErr: false,
		Check: healthPg.New(healthPg.Config{
			DSN: app.config.DB.DSN,
		}),
	})
	if err != nil {
		return errs.NewCommonError("failed to register pg health checker", err)
	}

	app.health = healthChecker

	return nil
}

func (app *App) initHTTPRouter() error {
	app.httpRouter = rest.NewAppChiRouter(app.config.HTTP, app.config.Telemetry, app.logger, app.health, nil, nil, app.testFacade)

	return nil
}

func (app *App) initHTTPServer() error {
	app.httpServer = &http.Server{
		Addr:    app.config.HTTP.Address,
		Handler: app.httpRouter.GetRouter(),
	}

	return nil
}

func (app *App) initGRPCService() error {
	app.grpcExampleService = grpcsvc.NewExampleGRPCService(app.config.GRPC, app.testFacade, app.logger)

	return nil
}

func (app *App) initGRPCServer() error {
	// Настраиваем KeepAlive на основе твоего GRPCConfig
	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     app.config.GRPC.MaxConnIdle,
		MaxConnectionAge:      app.config.GRPC.MaxConnAge,
		MaxConnectionAgeGrace: app.config.GRPC.MaxConnAgeGrace,
		Time:                  app.config.GRPC.KeepAliveTime,
		Timeout:               app.config.GRPC.KeepAliveTimeout,
	}
	// Метрики
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
		// Add tenant_name as a context label. This server option is necessary
		// to initialize the metrics with the labels that will be provided
		// dynamically from the context. This should be used in tandem with
		// WithLabelsFromContext in the interceptor options.
		grpcprom.WithContextLabels("tenant_name"),
	)
	if err := prometheus.Register(srvMetrics); err != nil {
		return errs.NewCommonError("failed to register grpc metrics", err)
	}
	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}
	// Extract the tenant name value from gRPC metadata
	// and use it as a label on our metrics.
	labelsFromContext := func(ctx context.Context) prometheus.Labels {
		labels := prometheus.Labels{}

		md := metadata.ExtractIncoming(ctx)
		tenantName := md.Get("tenant-name")
		if tenantName == "" {
			tenantName = "unknown"
		}
		labels["tenant_name"] = tenantName

		return labels
	}
	// Setup metric for panic recoveries.
	panicsTotal := promauto.NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})
	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		//		rpcLogger.Error("recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}

	// real IP
	realIPOpts := []realip.Option{
		realip.WithHeaders([]string{
			realip.XRealIp,
			realip.XForwardedFor,
			realip.TrueClientIp,
		}),
	}

	// Собираем опции сервера
	opts := []grpc.ServerOption{
		// keepalive
		grpc.KeepaliveParams(kasp),
		// timeout
		grpc.ConnectionTimeout(app.config.GRPC.Timeout),
		// tracing
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// metrics
		grpc.ChainUnaryInterceptor(
			srvMetrics.UnaryServerInterceptor(
				grpcprom.WithExemplarFromContext(exemplarFromContext),
				grpcprom.WithLabelsFromContext(labelsFromContext),
			),
			realip.UnaryServerInterceptorOpts(realIPOpts...),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			srvMetrics.StreamServerInterceptor(
				grpcprom.WithExemplarFromContext(exemplarFromContext),
				grpcprom.WithLabelsFromContext(labelsFromContext),
			),
			realip.StreamServerInterceptorOpts(realIPOpts...),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	}

	app.grpcServer = grpc.NewServer(opts...)

	// Регистрация
	pb.RegisterExampleServiceServer(app.grpcServer, app.grpcExampleService)

	// Инициализация метрик с нулевыми рядами
	srvMetrics.InitializeMetrics(app.grpcServer)

	return nil
}
