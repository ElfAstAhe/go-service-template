package azure

import (
	"context"
	"strings"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

type ConnectorOption func(*ConnectorOptions)

type ConnectorOptions struct {
	URL             string
	ConnOpts        *amqp.ConnOptions
	SessionOpts     *amqp.SessionOptions
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	DialFnTestGap   func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)
	Logger          logger.Logger
}

func NewConnectorOptions() *ConnectorOptions {
	return &ConnectorOptions{
		ConnectTimeout:  DefaultConnectTimeout,
		ShutdownTimeout: DefaultShutdownTimeout,
	}
}

func (co *ConnectorOptions) Validate() error {
	if strings.TrimSpace(co.URL) == "" {
		return errs.NewTlCommonError("Validate", "amqp connection URL cannot be empty", nil)
	}
	if co.Logger == nil {
		return errs.NewTlCommonError("Validate", "connector logger is nil", nil)
	}
	if co.ConnectTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "connection timeout is invalid", nil)
	}
	if co.ShutdownTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "shutdown timeout is invalid", nil)
	}

	return nil
}

func WithConnectorURL(url string) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.URL = url
	}
}

func WithConnectorConnectTimeout(timeout time.Duration) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.ConnectTimeout = timeout
	}
}

func WithConnectorShutdownTimeout(timeout time.Duration) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.ShutdownTimeout = timeout
	}
}

func WithConnectorDialFnTestGap(fn func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.DialFnTestGap = fn
	}
}

func WithConnectorLogger(log logger.Logger) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.Logger = log
	}
}

func WithConnectorConnOpts(connOpts *amqp.ConnOptions) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.ConnOpts = connOpts
	}
}

func WithConnectorSessionOpts(sessOpts *amqp.SessionOptions) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.SessionOpts = sessOpts
	}
}
