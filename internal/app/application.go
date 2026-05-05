package app

import (
	"errors"

	"github.com/ElfAstAhe/go-service-template/internal/app/container"
	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/app"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type Application struct {
	*app.BaseApplication
	conf *config.Config
	log  logger.Logger
}

func NewApplication(opts ...Option) *Application {
	// create instance
	res := &Application{}
	// setup
	for _, opt := range opts {
		opt(res)
	}
	// embed
	res.BaseApplication = app.NewBaseApplication(
		app.WithOrchestrator(container.NewOrchestrator()),
		app.WithLogger(res.log),
		app.WithCloseTimeout(res.conf.App.CloseTimeout),
		app.WithStopTimeout(res.conf.App.StopTimeout),
	)

	return res
}

func (app *Application) Init() error {
	var cntErrs []error
	appCnt := container.NewAppContainer(app.GetOrchestrator())
	// register containers
	cntErrs = append(cntErrs,
		app.GetOrchestrator().Register(appCnt),
		app.GetOrchestrator().Register(container.NewToolsContainer(app.GetOrchestrator())),
	)
	err := errors.Join(cntErrs...)
	if err != nil {
		return errs.NewCommonError("init application failed", err)
	}

	// register app params
	var regErrs []error

	regErrs = append(regErrs,
		appCnt.RegisterInstance(container.LoggerInstance, app.log),
		appCnt.RegisterInstance(container.ConfigInstance, app.conf),
	)
	err = errors.Join(regErrs...)
	if err != nil {
		return errs.NewCommonError("register application params failed", err)
	}

	return app.BaseApplication.Init()
}

func (app *Application) Run() error {
	// ToDo: implement

	return app.BaseApplication.Run()
}

func (app *Application) Stop() error {
	// ToDo: implement

	return app.BaseApplication.Stop()
}

func (app *Application) Close() error {
	// ToDo: implement

	return app.BaseApplication.Close()
}
