package azure

import (
	"context"
	"strings"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

const (
	DefaultReceiverConnectTimeout  time.Duration = 5 * time.Second
	DefaultReceiverShutdownTimeout time.Duration = 5 * time.Second
	DefaultReceiverCredit          int           = 50
)

type ReceiverOption func(*ClientReceiverOptions)

type ClientReceiverOptions struct {
	URL             string
	ConnOpts        *amqp.ConnOptions
	SessionOpts     *amqp.SessionOptions
	ReceiverOpts    *amqp.ReceiverOptions // Опции для Durable подписок ресивера
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	DialFnTestGap   func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)
	Logger          logger.Logger
}

func NewClientReceiverOptions() *ClientReceiverOptions {
	return &ClientReceiverOptions{
		URL:             config.DefaultAMQPSenderURL, // Поменяй на дефолт ресивера, если в конфиге есть разделение
		ConnectTimeout:  DefaultReceiverConnectTimeout,
		ShutdownTimeout: DefaultReceiverShutdownTimeout,
	}
}

func (cro *ClientReceiverOptions) Validate() error {
	if strings.TrimSpace(cro.URL) == "" {
		return errs.NewTlCommonError("Validate", "amqp URL empty", nil)
	}
	if cro.Logger == nil {
		return errs.NewTlCommonError("Validate", "logger is nil", nil)
	}
	if cro.ConnectTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "connection timeout is invalid", nil)
	}
	if cro.ShutdownTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "shutdown timeout is invalid", nil)
	}

	return nil
}

func WithReceiverURL(url string) ReceiverOption {
	return func(cro *ClientReceiverOptions) {
		cro.URL = url
	}
}

func WithReceiverConnectTimeout(timeout time.Duration) ReceiverOption {
	return func(cro *ClientReceiverOptions) {
		cro.ConnectTimeout = timeout
	}
}

func WithReceiverShutdownTimeout(timeout time.Duration) ReceiverOption {
	return func(cro *ClientReceiverOptions) {
		cro.ShutdownTimeout = timeout
	}
}

func WithReceiverDialFnTestGap(fn func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)) ReceiverOption {
	return func(cro *ClientReceiverOptions) {
		cro.DialFnTestGap = fn
	}
}

func WithReceiverLogger(log logger.Logger) ReceiverOption {
	return func(cro *ClientReceiverOptions) {
		cro.Logger = log
	}
}

func WithReceiverConnOpts(connOpts *amqp.ConnOptions) ReceiverOption {
	return func(cro *ClientReceiverOptions) {
		cro.ConnOpts = connOpts
	}
}

func WithReceiverSessionOpts(sessOpts *amqp.SessionOptions) ReceiverOption {
	return func(cro *ClientReceiverOptions) {
		cro.SessionOpts = sessOpts
	}
}

func WithReceiverOpts(receiverOpts *amqp.ReceiverOptions) ReceiverOption {
	return func(cro *ClientReceiverOptions) {
		cro.ReceiverOpts = receiverOpts
	}
}
