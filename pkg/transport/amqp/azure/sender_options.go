package azure

import (
	"strings"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
)

const (
	DefaultSenderConnectTimeout        time.Duration = 5 * time.Second
	DefaultSenderShutdownTimeout       time.Duration = 5 * time.Second
	DefaultSenderPublishMaxTryAttempts int           = 2
	DefaultSenderPublishBaseRetryDelay time.Duration = 100 * time.Millisecond
	DefaultSenderPublishMaxRetryDelay  time.Duration = 3 * time.Second
)

type SenderOption func(*SenderOptions)

type SenderOptions struct {
	Connector             pkgamqp.Connector[*amqp.Session]
	TargetName            string
	Opts                  *amqp.SenderOptions
	ConnectTimeout        time.Duration
	ShutdownTimeout       time.Duration
	Logger                logger.Logger
	PublishMaxTryAttempts int
	PublishBaseRetryDelay time.Duration
	PublishMaxRetryDelay  time.Duration
}

func NewSenderOptions() *SenderOptions {
	return &SenderOptions{
		ConnectTimeout:        DefaultSenderConnectTimeout,
		ShutdownTimeout:       DefaultSenderShutdownTimeout,
		PublishMaxTryAttempts: DefaultSenderPublishMaxTryAttempts,
		PublishBaseRetryDelay: DefaultSenderPublishBaseRetryDelay,
		PublishMaxRetryDelay:  DefaultSenderPublishMaxRetryDelay,
	}
}

func (so *SenderOptions) Validate() error {
	if so.Connector == nil {
		return errs.NewTlCommonError("Validate", "connector is required and cannot be nil", nil)
	}
	if strings.TrimSpace(so.TargetName) == "" {
		return errs.NewTlCommonError("Validate", "target name empty", nil)
	}
	if so.ConnectTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "connection timeout is invalid", nil)
	}
	if so.ShutdownTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "shutdown timeout is invalid", nil)
	}
	if so.PublishMaxTryAttempts <= 0 {
		return errs.NewTlCommonError("Validate", "publish max try attempts is invalid", nil)
	}
	if so.PublishBaseRetryDelay <= 0 {
		return errs.NewTlCommonError("Validate", "publish base retry delay is invalid", nil)
	}
	if so.PublishMaxRetryDelay <= 0 {
		return errs.NewTlCommonError("Validate", "publish max retry delay is invalid", nil)
	}
	if so.Logger == nil {
		return errs.NewTlCommonError("Validate", "logger is nil", nil)
	}
	if so.PublishBaseRetryDelay > so.PublishMaxRetryDelay {
		return errs.NewTlCommonError("Validate", "base retry delay cannot be greater than max retry delay", nil)
	}

	return nil
}

func WithSenderConnector(connector pkgamqp.Connector[*amqp.Session]) SenderOption {
	return func(so *SenderOptions) {
		so.Connector = connector
	}
}

func WithSenderTargetName(targetName string) SenderOption {
	return func(so *SenderOptions) {
		so.TargetName = targetName
	}
}

func WithSenderConnectTimeout(timeout time.Duration) SenderOption {
	return func(so *SenderOptions) {
		so.ConnectTimeout = timeout
	}
}

func WithSenderShutdownTimeout(timeout time.Duration) SenderOption {
	return func(so *SenderOptions) {
		so.ShutdownTimeout = timeout
	}
}

func WithSenderLogger(log logger.Logger) SenderOption {
	return func(so *SenderOptions) {
		so.Logger = log
	}
}

func WithSenderPublishMaxTryAttempts(maxTryAttempts int) SenderOption {
	return func(so *SenderOptions) {
		so.PublishMaxTryAttempts = maxTryAttempts
	}
}

func WithSenderPublishBaseRetryDelay(delay time.Duration) SenderOption {
	return func(so *SenderOptions) {
		so.PublishBaseRetryDelay = delay
	}
}

func WithSenderPublishMaxRetryDelay(delay time.Duration) SenderOption {
	return func(so *SenderOptions) {
		so.PublishMaxRetryDelay = delay
	}
}

func WithSenderOpts(senderOpts *amqp.SenderOptions) SenderOption {
	return func(so *SenderOptions) {
		so.Opts = senderOpts
	}
}
