package azure

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"
	"time"

	"github.com/Azure/go-amqp"
)

type Option func(*options)

type options struct {
	*amqp.ConnOptions
	connectTimeout  time.Duration
	shutdownTimeout time.Duration
	dialFnTestGap   func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)
}

func WithConnectTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.connectTimeout = timeout
	}
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

// WithTLS задает конфигурацию TLS для защищенного соединения.
// Внимание: метод полностью перезаписывает структуру TLSConfig.
// Если вам нужно переопределить флаг InsecureSkipVerify,
// вызывайте опцию WithInsecureSkipVerify строго ПОСЛЕ WithTLS.
func WithTLS(config *tls.Config) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		o.ConnOptions.TLSConfig = config
	}
}

func WithInsecureSkipVerify(skip bool) Option {
	return func(o *options) {
		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		if o.ConnOptions.TLSConfig == nil {
			o.ConnOptions.TLSConfig = &tls.Config{}
		}
		o.ConnOptions.TLSConfig.InsecureSkipVerify = skip
	}
}

// WithCACerts загружает корневые сертификаты из одного или нескольких PEM-файлов.
func WithCACerts(certPaths ...string) Option {
	return func(o *options) {
		if len(certPaths) == 0 {
			return
		}

		if o.ConnOptions == nil {
			o.ConnOptions = &amqp.ConnOptions{}
		}
		if o.ConnOptions.TLSConfig == nil {
			o.ConnOptions.TLSConfig = &tls.Config{}
		}

		caCertPool, err := x509.SystemCertPool() // Загружает дефолтные сертификаты ОС
		if err != nil {
			caCertPool = x509.NewCertPool() // Фолбек, если в системе нет пула (например, в scratch-контейнере)
		}
		var loadedAny bool

		for _, path := range certPaths {
			if strings.TrimSpace(path) == "" {
				continue
			}

			certBytes, err := os.ReadFile(path)
			if err != nil {
				continue // Или логируем ошибку чтения конкретного файла
			}

			// AppendCertsFromPEM проглотит как один сертификат, так и всю цепочку в файле
			if ok := caCertPool.AppendCertsFromPEM(certBytes); ok {
				loadedAny = true
			}
		}

		// Если не удалось загрузить вообще ни одного сертификата из переданных путей,
		// инвалидируем пул, чтобы Dial упал, защищая периметр.
		if !loadedAny {
			o.ConnOptions.TLSConfig.RootCAs = x509.NewCertPool()
			return
		}

		o.ConnOptions.TLSConfig.RootCAs = caCertPool
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
