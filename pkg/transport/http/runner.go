package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type ServerProvider func(router Router, conf *config.HTTPConfig) (*http.Server, error)
type ServerLauncher func(server *http.Server, conf *config.HTTPConfig) error

type Runner struct {
	name           string
	router         Router
	server         *http.Server
	conf           *config.HTTPConfig
	running        *atomic.Bool
	serverProvider ServerProvider
	serverLauncher ServerLauncher
	log            logger.Logger
}

var _ container.Runner = (*Runner)(nil)

func NewRunner(
	opts ...Option,
) (*Runner, error) {
	res := &Runner{
		name:    "default http",
		conf:    config.NewDefaultHTTPConfig(),
		running: new(atomic.Bool),
	}
	res.running.Store(false)

	// apply options
	for _, option := range opts {
		option(res)
	}
	// check required
	if utils.IsNil(res.router) {
		return nil, errs.NewTlCommonError("NewRunner", "router not applied", nil)
	}
	if utils.IsNil(res.log) {
		return nil, errs.NewTlCommonError("NewRunner", "logger not applied", nil)
	}

	// check defaults
	if utils.IsNil(res.serverProvider) {
		res.serverProvider = res.defaultServerProvider
	}
	if utils.IsNil(res.serverLauncher) {
		res.serverLauncher = res.defaultServerLauncher
	}

	return res, nil
}

// Start create and then start http server
//
//	Attention! This method os blocked!
func (r *Runner) Start(ctx context.Context) error {
	r.log.Debugf("Runner.Start %s start", r.GetName())
	defer r.log.Debugf("Runner.Start %s finish", r.GetName())

	// switch running flag
	if !r.running.CompareAndSwap(false, true) {
		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s already started", r.GetName()), nil)
	}
	var err error
	// create server instance
	r.server, err = r.serverProvider(r.router, r.conf)
	if err != nil {
		r.running.Store(false)

		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s create http server failed", r.GetName()), err)
	}
	// launch server
	err = r.serverLauncher(r.server, r.conf)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		r.running.Store(false)

		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s http server listen failed", r.GetName()), err)
	}

	return nil
}

func (r *Runner) Stop(stopCtx context.Context) error {
	r.log.Debugf("Runner.Stop %s start", r.GetName())
	defer r.log.Debugf("Runner.Stop %s finish", r.GetName())

	if !r.running.CompareAndSwap(true, false) {
		return errs.NewTlCommonError("Runner.Stop", fmt.Sprintf("runner %s not running", r.GetName()), nil)
	}

	var srv *http.Server
	srv, r.server = r.server, nil

	shutdownCtx, shutdownCancel := context.WithTimeout(stopCtx, r.conf.ShutdownTimeout)
	defer shutdownCancel()

	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			r.log.Warn("http server shutdown timed out (force close)")
		} else {
			r.log.Errorf("http server shutdown with error [%v]", err)
		}

		return errs.NewTlCommonError("Runner.Stop", "http server shutdown with error", err)
	}
	r.log.Debug("http server shutdown gracefully complete")

	return nil
}

func (r *Runner) GetName() string {
	return r.name
}

func (r *Runner) IsRunning() bool {
	return r.running.Load()
}

func (r *Runner) defaultServerProvider(router Router, conf *config.HTTPConfig) (*http.Server, error) {
	r.log.Debugf("Runner.defaultServerProvider %s start", r.GetName())
	defer r.log.Debugf("Runner.defaultServerProvider %s finish", r.GetName())

	return &http.Server{
		Addr:         conf.Address,
		Handler:      router.GetRouter(),
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		IdleTimeout:  conf.IdleTimeout,
	}, nil
}

func (r *Runner) defaultServerLauncher(server *http.Server, conf *config.HTTPConfig) error {
	r.log.Debugf("Runner.defaultServerLauncher %s start", r.GetName())
	defer r.log.Debugf("Runner.defaultServerLauncher %s finish", r.GetName())

	if conf.Secure {
		r.log.Infof("Runner.defaultServerLauncher %s http secure listen %s", r.GetName(), conf.Address)

		return server.ListenAndServeTLS(conf.CertificatePath, conf.PrivateKeyPath)
	}

	r.log.Infof("Runner.defaultServerLauncher %s http nonsecure listen %s", r.GetName(), conf.Address)

	return server.ListenAndServe()
}
