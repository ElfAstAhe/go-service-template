package azure

import (
	"context"
	"errors"
	"sync"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type Connector struct {
	opts   *ConnectorOptions
	mu     sync.RWMutex
	initMu sync.Mutex
	conn   *amqp.Conn
	sess   *amqp.Session
	logger logger.Logger
}

var _ pkgamqp.Connector[*amqp.Session] = (*Connector)(nil)

func NewConnector(opts ...ConnectorOption) (*Connector, error) {
	cOpts := NewConnectorOptions()
	for _, opt := range opts {
		opt(cOpts)
	}
	err := cOpts.Validate()
	if err != nil {
		return nil, errs.NewTlCommonError("NewConnector", "connector options validation failed", err)
	}

	return &Connector{
		opts:   cOpts,
		logger: cOpts.Logger.GetLogger("azure-connector"),
		conn:   nil,
		sess:   nil,
	}, nil
}

// GetConnection — главная точка входа для всех сендеров и ресиверов.
// Возвращает живую общую сессию. Если связь порвана, атомарно восстановит её.
//
//goland:noinspection DuplicatedCode
func (c *Connector) GetConnection(ctx context.Context) (*amqp.Session, error) {
	// 1. Быстрый путь (Fast Path): если всё живо, отдаем под RLock за наносекунды
	c.mu.RLock()
	connAlive := !utils.IsNil(c.conn)
	sessAlive := !utils.IsNil(c.sess)
	localSess := c.sess
	c.mu.RUnlock()

	if connAlive && sessAlive {
		return localSess, nil
	}

	// 2. Медленный путь (Slow Path): сети нет.
	// Блокируем initMu, чтобы только ОДНА горутина пошла восстанавливать связь
	c.initMu.Lock()
	defer c.initMu.Unlock()

	// 3. Double-Check: возможно, пока мы ждали initMu, параллельный поток уже всё поднял
	c.mu.RLock()
	connAlive = !utils.IsNil(c.conn)
	sessAlive = !utils.IsNil(c.sess)
	localSess = c.sess
	c.mu.RUnlock()

	if connAlive && sessAlive {
		return localSess, nil
	}

	var newlyDialed bool
	localConn := c.conn

	// 4. Если физический сокет мертв — переподключаемся (сетевой вызов ВНЕ c.mu)
	if !connAlive {
		dialCtx, cancel := context.WithTimeout(ctx, c.opts.ConnectTimeout)
		defer cancel()

		var conn *amqp.Conn
		var err error
		if !utils.IsNil(c.opts.DialFnTestGap) {
			conn, err = c.opts.DialFnTestGap(dialCtx, c.opts.URL, c.opts.ConnOpts)
		} else {
			conn, err = amqp.Dial(dialCtx, c.opts.URL, c.opts.ConnOpts)
		}
		if err != nil {
			return nil, errs.NewTlCommonError("GetConnection", "amqp infrastructure dial failed", err)
		}
		localConn = conn
		newlyDialed = true
	}

	// 5. Создаем новую общую сессию (сетевой вызов ВНЕ c.mu)
	sessCtx, cancelSess := context.WithTimeout(ctx, c.opts.ConnectTimeout)
	defer cancelSess()

	session, err := localConn.NewSession(sessCtx, c.opts.SessionOpts)
	if err != nil {
		// Асинхронный сброс: обнуляем переменные за наносекунды под Lock,
		// а тяжелый Close() уводим в фон, чтобы не вешать бизнес-горутины
		c.mu.Lock()
		var connToCloseBehindMutex *amqp.Conn

		if newlyDialed {
			connToCloseBehindMutex = localConn
		} else {
			connToCloseBehindMutex = localConn
			c.conn = nil
		}
		c.sess = nil
		c.mu.Unlock()

		if connToCloseBehindMutex != nil {
			go func(cn *amqp.Conn) {
				_ = cn.Close()
			}(connToCloseBehindMutex)
		}

		return nil, errs.NewTlCommonError("GetConnection", "failed to open logical amqp session", err)
	}

	// 6. Успех — фиксируем новые живые ресурсы в структуре
	c.mu.Lock()
	c.conn = localConn
	c.sess = session
	c.mu.Unlock()

	return session, nil
}

// Invalidate анализирует сетевую ошибку Go-AMQP и сбрасывает соответствующий уровень ресурсов.
func (c *Connector) Invalidate(err error) {
	if err == nil {
		return
	}

	var sessionErr *amqp.SessionError
	var connErr *amqp.ConnError

	c.mu.Lock()
	defer c.mu.Unlock()

	switch {
	case errors.As(err, &connErr):
		c.logger.Error("AMQP physical connection socket dead. Invalidating full pipeline.")
		c.conn = nil
		c.sess = nil

	case errors.As(err, &sessionErr):
		c.logger.Warn("AMQP logical session dead. Invalidating session layer.")
		c.sess = nil
	}
}

// Close мягко закрывает общую сессию и физическое соединение всего микросервиса при завершении работы приложения.
func (c *Connector) Close(ctx context.Context) error {
	c.logger.Debug("connector shutdown started")
	defer c.logger.Debug("connector shutdown finished")

	c.mu.Lock()
	sessionToClose := c.sess
	connToClose := c.conn

	c.sess = nil
	c.conn = nil
	c.mu.Unlock() // МЬЮТЕКС МГНОВЕННО СВОБОДЕН

	closeCtx, closeCancel := context.WithTimeout(ctx, c.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)

	go func() {
		var closeErrs []error

		if !utils.IsNil(sessionToClose) {
			if err := sessionToClose.Close(closeCtx); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}

		if !utils.IsNil(connToClose) {
			if err := connToClose.Close(); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}

		done <- errors.Join(closeErrs...)
	}()

	select {
	case err := <-done:
		if err != nil {
			return errs.NewTlCommonError("Close", "Azure AMQP connector close fails", err)
		}

		return nil
	case <-closeCtx.Done():
		// Предохранитель: если библиотека зависла на Close, рвем сокет жестко на уровне ядра ОС
		if connToClose != nil {
			_ = connToClose.Close()
		}

		return errs.NewTlCommonError("Close", "Azure AMQP connector close timeout: connection force closed", closeCtx.Err())
	}
}
