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
	DefaultSenderConnectTimeout        time.Duration = 5 * time.Second
	DefaultSenderShutdownTimeout       time.Duration = 5 * time.Second
	DefaultSenderPublishMaxTryAttempts int           = 2
	DefaultSenderPublishBaseRetryDelay time.Duration = 100 * time.Millisecond
	DefaultSenderPublishMaxRetryDelay  time.Duration = 3 * time.Second
)

type SenderOption func(*ClientSenderOptions)

type ClientSenderOptions struct {
	URL                   string
	TargetName            string
	ConnOpts              *amqp.ConnOptions
	SessionOpts           *amqp.SessionOptions
	SenderOpts            *amqp.SenderOptions
	ConnectTimeout        time.Duration
	ShutdownTimeout       time.Duration
	DialFnTestGap         func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)
	Logger                logger.Logger
	PublishMaxTryAttempts int
	PublishBaseRetryDelay time.Duration
	PublishMaxRetryDelay  time.Duration
}

func NewClientSenderOptions() *ClientSenderOptions {
	return &ClientSenderOptions{
		URL:                   config.DefaultAMQPSenderURL,
		ConnectTimeout:        DefaultSenderConnectTimeout,
		ShutdownTimeout:       DefaultSenderShutdownTimeout,
		PublishMaxTryAttempts: DefaultSenderPublishMaxTryAttempts,
		PublishBaseRetryDelay: DefaultSenderPublishBaseRetryDelay,
		PublishMaxRetryDelay:  DefaultSenderPublishMaxRetryDelay,
	}
}

func (cso *ClientSenderOptions) Validate() error {
	if strings.TrimSpace(cso.URL) == "" {
		return errs.NewTlCommonError("Validate", "amqp URL empty", nil)
	}
	//if strings.TrimSpace(cso.TargetName) == "" {
	//	return errs.NewTlCommonError("Validate", "target name empty", nil)
	//}
	//if utils.IsNil(cso.ConnOpts) {
	//    return errs.NewTlCommonError("Validate", "connection options is nil", nil)
	//}
	//if utils.IsNil(cso.SessionOpts) {
	//    return errs.NewTlCommonError("Validate", "session options is nil", nil)
	//}
	if !(cso.ConnectTimeout > 0) {
		return errs.NewTlCommonError("Validate", "connection timeout is invalid", nil)
	}
	if !(cso.ShutdownTimeout > 0) {
		return errs.NewTlCommonError("Validate", "shutdown timeout is invalid", nil)
	}
	if !(cso.PublishMaxTryAttempts > 0) {
		return errs.NewTlCommonError("Validate", "publish max try attempts is invalid", nil)
	}
	if !(cso.PublishBaseRetryDelay > 0) {
		return errs.NewTlCommonError("Validate", "publish base retry delay is invalid", nil)
	}
	if !(cso.PublishMaxRetryDelay > 0) {
		return errs.NewTlCommonError("Validate", "publish max retry delay is invalid", nil)
	}
	if cso.Logger == nil {
		return errs.NewTlCommonError("Validate", "logger is nil", nil)
	}
	if cso.PublishBaseRetryDelay > cso.PublishMaxRetryDelay {
		return errs.NewTlCommonError("Validate", "base retry delay cannot be greater than max retry delay", nil)
	}

	return nil
}

func WithSenderURL(url string) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.URL = url
	}
}

func WithSenderTargetName(targetName string) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.TargetName = targetName
	}
}

func WithSenderConnectTimeout(timeout time.Duration) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.ConnectTimeout = timeout
	}
}

func WithSenderShutdownTimeout(timeout time.Duration) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.ShutdownTimeout = timeout
	}
}

func WithSenderDialFnTestGap(fn func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.DialFnTestGap = fn
	}
}

func WithSenderLogger(log logger.Logger) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.Logger = log
	}
}

func WithSenderPublishMaxTryAttempts(maxTryAttempts int) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.PublishMaxTryAttempts = maxTryAttempts
	}
}

func WithSenderPublishBaseRetryDelay(delay time.Duration) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.PublishBaseRetryDelay = delay
	}
}

func WithSenderPublishMaxRetryDelay(delay time.Duration) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.PublishMaxRetryDelay = delay
	}
}

func WithSenderConnOpts(connOpts *amqp.ConnOptions) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.ConnOpts = connOpts
	}
}

func WithSenderSessionOpts(sessOpts *amqp.SessionOptions) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.SessionOpts = sessOpts
	}
}

func WithSenderOpts(senderOpts *amqp.SenderOptions) SenderOption {
	return func(cso *ClientSenderOptions) {
		cso.SenderOpts = senderOpts
	}
}
