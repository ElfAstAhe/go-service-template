package kafka

import (
	"context"
	"net"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type ConnectorOption func(*ConnectorOptions)

type ConnectorOptions struct {
	Brokers         []string         // Список адресов брокеров кластера Kafka (например, ["localhost:9092"])
	Dialer          *kafka.Dialer    // Кастомный нативный диалер для тонкой настройки TCP/TLS сокетов
	Transport       *kafka.Transport // Настройки транспортного слоя (пулы коннектов, TLS, SASL)
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	DialFnTestGap   func(ctx context.Context, network, addr string) (net.Conn, error) // Тестовая заглушка для мока сокетов
	Logger          logger.Logger
}

func NewConnectorOptions() *ConnectorOptions {
	return &ConnectorOptions{
		ConnectTimeout:  DefaultConnectTimeout,
		ShutdownTimeout: DefaultShutdownTimeout,
	}
}

func (co *ConnectorOptions) Validate() error {
	if len(co.Brokers) == 0 {
		return errs.NewTlCommonError("Validate", "kafka bootstrap brokers list cannot be empty", nil)
	}
	for _, broker := range co.Brokers {
		if broker == "" {
			return errs.NewTlCommonError("Validate", "kafka broker address cannot be empty string", nil)
		}
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

func WithConnectorBrokers(brokers []string) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.Brokers = brokers
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

func WithConnectorDialFnTestGap(fn func(ctx context.Context, network, addr string) (net.Conn, error)) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.DialFnTestGap = fn
	}
}

func WithConnectorLogger(log logger.Logger) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.Logger = log
	}
}

func WithConnectorDialer(dialer *kafka.Dialer) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.Dialer = dialer
	}
}

func WithConnectorTransport(transport *kafka.Transport) ConnectorOption {
	return func(co *ConnectorOptions) {
		co.Transport = transport
	}
}
