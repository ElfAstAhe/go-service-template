package azure

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/Azure/go-amqp"
)

type Option func(*options)

type options struct {
	*amqp.ConnOptions
	shutdownTimeout time.Duration
	dialFnTestGap   func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.shutdownTimeout = timeout
	}
}

func WithSASLPlain(username, password string) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.SASLType = amqp.SASLTypePlain(username, password)
	}
}

func WithTLS(config *tls.Config) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.TLSConfig = config
	}
}

func WithContainerID(id string) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.ContainerID = id
	}
}

func WithHostName(hostname string) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.HostName = hostname
	}
}

func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.IdleTimeout = idleTimeout
	}
}

func WithMaxFrameSize(maxFrameSize uint32) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.MaxFrameSize = maxFrameSize
	}
}

func WithMaxSessions(maxSessions uint16) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.MaxSessions = maxSessions
	}
}

func WithProps(props map[string]any) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.Properties = props
	}
}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.WriteTimeout = writeTimeout
	}
}

func WithDialFnTestGap(fn func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)) Option {
	return func(o *options) {
		o.dialFnTestGap = fn
	}
}
