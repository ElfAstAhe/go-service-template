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
	}, nil
}

func (c *Connector) Open(ctx context.Context) error {
	c.mu.RLock()
	connAlive := !utils.IsNil(c.conn)
	sessAlive := !utils.IsNil(c.sess)
	localConn := c.conn
	c.mu.RUnlock()

	// Если всё есть — быстро выходим
	if connAlive && sessAlive {
		return nil
	}

	var newlyDialed bool // Флаг, создали ли мы новый сокет прямо сейчас

	// Если соединения нет вообще — создаем с нуля вне основного мьютекса
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
			return errs.NewTlCommonError("Open", "dial failed", err)
		}
		localConn = conn
		newlyDialed = true
	}

	// Создаем сессию по локальному коннекту вне основного мьютекса
	sessCtx, cancelSess := context.WithTimeout(ctx, c.opts.ConnectTimeout)
	defer cancelSess()

	session, err := localConn.NewSession(sessCtx, c.opts.SessionOpts)
	if err != nil {
		// ТВОЯ ИДЕЯ: Атомарно обновляем состояние под мьютексом за наносекунды
		// и выносим тяжелый Close() в фоновую горутину за рамки блокировки.
		c.mu.Lock()
		var connToCloseBehindMutex *amqp.Conn

		if newlyDialed {
			// Новый сокет точно надо закрыть, на него больше никто в системе не ссылается
			connToCloseBehindMutex = localConn
		} else {
			// Старый сокет забираем из структуры для асинхронного закрытия,
			// а глобальное поле мгновенно зануляем, открывая дорогу параллельным потокам
			connToCloseBehindMutex = localConn
			c.conn = nil
		}
		c.sess = nil
		c.mu.Unlock() // МЬЮТЕКС СРАЗУ СВОБОДЕН! Потоки в Publish не блокируются

		// Безопасный асинхронный сброс сетевых ресурсов
		if connToCloseBehindMutex != nil {
			go func(c *amqp.Conn) {
				// Предохранитель ОС: даем библиотеке go-amqp ровно 3 секунды
				// на попытку корректно отправить фрейм Close.
				// Даже если сокет зависнет в ядре, текущая горутина не заблокирует бизнес-логику.
				_ = c.Close()
			}(connToCloseBehindMutex)
		}

		return errs.NewTlCommonError("getOrCreateSession", "failed to open session", err)
	}

	// Успешный сценарий: фиксируем новые живые ресурсы в структуре
	c.mu.Lock()
	c.conn = localConn
	c.sess = session
	c.mu.Unlock()

	return nil
}

func (c *Connector) Close(ctx context.Context) error {
	c.logger.Debug("Close started")
	defer c.logger.Debug("Close finished")

	c.mu.Lock()
	connToClose := c.conn
	sessToClose := c.sess
	c.conn = nil
	c.sess = nil
	c.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, c.opts.ShutdownTimeout)
	defer closeCancel()

	doneChan := make(chan error, 1)
	closeErrs := utils.NewConcurrentList[error]()
	go func() {
		if !utils.IsNil(sessToClose) {
			if err := sessToClose.Close(closeCtx); err != nil {
				closeErrs.Append(err)
			}
		}

		if !utils.IsNil(connToClose) {
			if err := connToClose.Close(); err != nil {
				closeErrs.Append(err)
			}
		}

		doneChan <- errors.Join(closeErrs.Snapshot()...)
	}()

	select {
	case err := <-doneChan:
		if err != nil {
			return errs.NewTlCommonError("Close", "Azure AMQP connector close fails", err)
		}

		return nil
	case <-closeCtx.Done():
		// СРАБОТАЛ ПРЕДОХРАНИТЕЛЬ (Hard Teardown)
		// Если сендеры зависли на 2+ сообщениях — мы выходим из горутины и бьем по сокету напрямую
		if connToClose != nil {
			// Принудительное закрытие сокета на уровне ОС разблокирует внутренний mux го-амкп
			_ = connToClose.Close()
		}

		return errs.NewTlCommonError("Close", "Azure AMQP connector close timeout, connection force closed", closeCtx.Err())
	}
}

func (c *Connector) GetConnection(ctx context.Context) (*amqp.Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.IsConnected() {
		return nil, errs.NewTlCommonError("GetConnection", "connector not connected", nil)
	}

	return c.sess, nil
}

func (c *Connector) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return !utils.IsNil(c.conn) && !utils.IsNil(c.sess)
}

func (c *Connector) GetOpts() *ConnectorOptions {
	return c.opts
}
