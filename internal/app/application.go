package app

import (
	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/pkg/app"
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
		app.WithOrchestrator(NewOrchestrator()),
		app.WithLogger(res.log),
		app.WithCloseTimeout(res.conf.App.CloseTimeout),
		app.WithStopTimeout(res.conf.App.StopTimeout),
	)

	return res
}

func (app *Application) Init() error {
	// ToDo: implement

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
