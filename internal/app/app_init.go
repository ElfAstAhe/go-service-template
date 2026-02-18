package app

import (
	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/hellofresh/health-go/v5"
	healthPg "github.com/hellofresh/health-go/v5/checks/pgx5"
)

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

func (app *App) initRouter() {}
