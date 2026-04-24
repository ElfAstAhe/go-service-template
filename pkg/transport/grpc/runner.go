package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/grpc/interceptors"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ServerProvider func(conf *config.GRPCConfig) (*grpc.Server, error)
type ServiceRegister func(server *grpc.Server) error
type ServerLauncher func(server *grpc.Server, conf *config.GRPCConfig) error

type Runner struct {
	name            string
	server          *grpc.Server
	conf            *config.GRPCConfig
	running         *atomic.Bool
	serverProvider  ServerProvider
	serviceRegister ServiceRegister
	serverLauncher  ServerLauncher
	log             logger.Logger
	env             config.AppEnv
}

var _ container.Runner = (*Runner)(nil)

func NewRunner(opts ...Option) (*Runner, error) {
	// new instance
	res := &Runner{
		name:    "default gRPC",
		conf:    config.NewDefaultGRPCConfig(),
		running: new(atomic.Bool),
		env:     config.AppEnvProduction,
	}
	res.running.Store(false)

	// apply options
	for _, option := range opts {
		option(res)
	}

	// check required
	if utils.IsNil(res.serviceRegister) {
		return nil, errs.NewTlCommonError("NewRunner", "service register function not applied", nil)
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

// Start create, register and then start gRPC server
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
	r.server, err = r.serverProvider(r.conf)
	if err != nil {
		r.running.Store(false)

		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s create gRPC server failed", r.GetName()), err)
	}
	// register services
	err = r.serviceRegister(r.server)
	if err != nil {
		r.running.Store(false)

		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s register gRPC services failed", r.GetName()), err)
	}

	// launch server
	err = r.serverLauncher(r.server, r.conf)
	if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		r.running.Store(false)

		return errs.NewTlCommonError("Runner.Start", fmt.Sprintf("runner %s gRPC server listen failed", r.GetName()), err)
	}

	return nil
}

func (r *Runner) Stop(stopCtx context.Context) error {
	r.log.Debugf("Runner.Stop %s start", r.GetName())
	defer r.log.Debugf("Runner.Stop %s finish", r.GetName())

	if !r.running.CompareAndSwap(true, false) {
		return errs.NewTlCommonError("Runner.Stop", fmt.Sprintf("runner %s not running", r.GetName()), nil)
	}

	var srv *grpc.Server
	srv, r.server = r.server, nil

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		srv.GracefulStop()
	}()
	select {
	case <-doneChan:
		r.log.Info("gRPC server shutdown gracefully complete")
	case <-stopCtx.Done():
		r.log.Info("gRPC server shutdown main application context timed out")
		srv.Stop()

		return errs.NewTlCommonError("Runner.Stop", "gRPC server shutdown main application context timed out (force close)", stopCtx.Err())
	case <-time.After(r.conf.ShutdownTimeout):
		r.log.Info("gRPC server shutdown gracefully timeout")
		srv.Stop()

		return errs.NewTlCommonError("Runner.Stop", "gRPC server shutdown timed out (force close)", nil)
	}

	return nil
}

func (r *Runner) IsRunning() bool {
	return r.running.Load()
}

func (r *Runner) GetName() string {
	return r.name
}

func (r *Runner) defaultServerProvider(conf *config.GRPCConfig) (*grpc.Server, error) {
	// Настраиваем KeepAlive на основе твоего GRPCConfig
	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     conf.MaxConnIdle,
		MaxConnectionAge:      conf.MaxConnAge,
		MaxConnectionAgeGrace: conf.MaxConnAgeGrace,
		Time:                  conf.KeepAliveTime,
		Timeout:               conf.KeepAliveTimeout,
	}
	grpcPanicRecoveryHandler := func(p any) (err error) {
		r.log.Error("recovered from panic", "panic", p, "stack", debug.Stack())

		return status.Errorf(codes.Internal, "%s", p)
	}

	// Собираем опции сервера
	opts := []grpc.ServerOption{
		// keepalive
		grpc.KeepaliveParams(kasp),
		// timeout
		grpc.ConnectionTimeout(conf.Timeout),
		// tracing
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// metrics
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDExtractorUSInterceptor([]string{
				interceptors.MDXRequestID,
				interceptors.MDXCorrelationID,
				interceptors.MDRequestID,
			}),
			interceptors.TraceIDExtractorUSInterceptor([]string{
				interceptors.MDXCloudTraceContext,
				interceptors.MDTraceParent,
				interceptors.MDXTraceID,
				interceptors.MDTraceID,
			}),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			interceptors.RequestIDExtractorSSInterceptor([]string{
				interceptors.MDXRequestID,
				interceptors.MDXCorrelationID,
				interceptors.MDRequestID,
			}),
			interceptors.TraceIDExtractorSSInterceptor([]string{
				interceptors.MDXCloudTraceContext,
				interceptors.MDTraceParent,
				interceptors.MDXTraceID,
				interceptors.MDTraceID,
			}),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	}

	srv := grpc.NewServer(opts...)

	if r.env != config.AppEnvProduction {
		reflection.Register(srv)
	}

	return srv, nil
}

func (r *Runner) defaultServerLauncher(server *grpc.Server, conf *config.GRPCConfig) error {
	r.log.Debugf("Runner.defaultServerLauncher %s start", r.GetName())
	defer r.log.Debugf("Runner.defaultServerLauncher %s finish", r.GetName())

	lis, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return errs.NewTlCommonError("Runner.defaultServerLauncher", fmt.Sprintf("failed to listen %s", conf.Address), err)
	}

	r.log.Infof("Runner.defaultServerLauncher %s gRPC listen %s", r.GetName(), conf.Address)

	return server.Serve(lis)
}
