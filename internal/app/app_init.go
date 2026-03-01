package app

import (
	"net/http"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/facade"
	"github.com/ElfAstAhe/go-service-template/internal/repository"
	"github.com/ElfAstAhe/go-service-template/internal/repository/postgres"
	"github.com/ElfAstAhe/go-service-template/internal/transport/rest"
	"github.com/ElfAstAhe/go-service-template/internal/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	migrations "github.com/ElfAstAhe/go-service-template/pkg/migration/goose"
	"github.com/hellofresh/health-go/v5"
	healthPg "github.com/hellofresh/health-go/v5/checks/pgx5"
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
	app.httpRouter = rest.NewAppChiRouter(app.config.HTTP, app.logger, app.health, nil, nil, app.testFacade)

	return nil
}

func (app *App) initHTTPServer() error {
	app.httpServer = &http.Server{
		Addr:    app.config.HTTP.Address,
		Handler: app.httpRouter.GetRouter(),
	}

	return nil
}

func (app *App) initGRPCServer() error {
	// ToDo: implement

	return errs.NewNotImplementedError(nil)
}
