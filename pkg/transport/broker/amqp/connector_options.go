package amqp

import (
	"context"
	"strings"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
)

// Константы рантайм-дефолтов для внутренней защиты глобального коннектора Azure Service Bus.
const (
	defaultConnectTimeout  time.Duration = 5 * time.Second
	defaultShutdownTimeout time.Duration = 5 * time.Second
)

// ConnectorOption определяет функциональный тип для конфигурации опций коннектора (Fluent API).
type ConnectorOption func(*ConnectorOptions)

// ConnectorOptions инкапсулирует параметры рантайма, необходимые для сборки и жизненного цикла Connector.
type ConnectorOptions struct {
	// URL хранит строку подключения к брокеру (например, "amqps://RootManageSharedAccessKey:pass@ns.servicebus.windows.net").
	URL string
	// ConnOpts содержит специфичные низкоуровневые настройки сетевого соединения библиотеки Azure go-amqp.
	ConnOpts *amqp.ConnOptions
	// SessionOpts инкапсулирует параметры для создания общей AMQP-сессии поверх физического коннекта.
	SessionOpts     *amqp.SessionOptions
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	// DialFnTestGap — инъекция mock-функции дозвона для полной изоляции сетевого слоя в юнит-тестах.
	DialFnTestGap func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error)
	Logger        logger.Logger
}

// NewConnectorOptions создает структуру опций, сразу наполняя её безопасными сетевыми дефолтами.
func NewConnectorOptions() *ConnectorOptions {
	return &ConnectorOptions{
		ConnectTimeout:  defaultConnectTimeout,
		ShutdownTimeout: defaultShutdownTimeout,
	}
}

// Validate выполняет строгую проверку входящих параметров рантайма перед созданием общего коннектора.
// Защищает приложение от паник при парсинге пустой строки подключения или отсутствии логгера.
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

// ====================================================================
// Fluent API методы конфигурации
// ====================================================================

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
