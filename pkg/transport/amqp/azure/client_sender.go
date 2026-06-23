package azure

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// amqpSenderLink описывает методы встроенного отправителя библиотеки Azure AMQP,
// которые нам нужны для управления его жизненным циклом.
type amqpSenderLink interface {
	Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error
	Close(ctx context.Context) error
}

type ClientSender struct {
	mu      sync.RWMutex
	url     string
	logger  logger.Logger
	conn    *amqp.Conn
	session *amqp.Session
	senders map[string]amqpSenderLink
	opts    *options
}

var _ pkgamqp.ClientSender = (*ClientSender)(nil)

func NewClientSender(url string, log logger.Logger, opts ...Option) *ClientSender {
	conf := &options{
		ConnOptions: &amqp.ConnOptions{},
	}

	for _, opt := range opts {
		opt(conf)
	}

	return &ClientSender{
		url:     url,
		logger:  log.GetLogger("azure-client-sender"),
		senders: make(map[string]amqpSenderLink),
		opts:    conf,
	}
}

//goland:noinspection DuplicatedCode
func (cs *ClientSender) Close(ctx context.Context) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, cs.opts.shutdownTimeout)
	defer closeCancel()

	var closeErrs []error

	for addr, sender := range cs.senders {
		err := sender.Close(closeCtx)
		if err != nil {
			closeErrs = append(closeErrs, err)
		}
		delete(cs.senders, addr)
	}

	if cs.session != nil {
		err := cs.session.Close(closeCtx)
		if err != nil {
			closeErrs = append(closeErrs, err)
		}
		cs.session = nil
	}

	if cs.conn != nil {
		err := cs.conn.Close()
		if err != nil {
			closeErrs = append(closeErrs, err)
		}
		cs.conn = nil
	}

	err := errors.Join(closeErrs...)
	if err != nil {
		return errs.NewCommonError("Azure AMQP client sender close fails", err)
	}

	return nil
}

// Publish отправляет сообщение с автоматическим фоновым In-Flight реконнектом и ретраем
func (cs *ClientSender) Publish(ctx context.Context, targetName string, msg *pkgamqp.Message) error {
	if utils.IsNil(msg) {
		return errs.NewCommonError("cannot publish nil message", nil)
	}

	const maxAttempts = 2

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Лениво берем или открываем соединение прямо в момент отправки
		sender, err := cs.getOrCreateSender(ctx, targetName)
		if err != nil {
			if attempt == 1 {
				cs.logger.Warnf("AMQP failed to get sender on attempt 1, resetting session for retry: %v", err)
				cs.resetInfrastructure()
				continue
			}

			return errs.NewCommonError("azure sender failed to initialize connection", err)
		}

		azureMsg := amqp.NewMessage(msg.Payload)
		azureMsg.Properties = &amqp.MessageProperties{
			ContentType: cs.pStr("application/json"),
		}

		if len(msg.Properties) > 0 {
			azureMsg.ApplicationProperties = msg.Properties
		}

		// Попытка отправить пакет в сеть
		err = sender.Send(ctx, azureMsg, nil)
		if err == nil {
			if attempt > 1 {
				cs.logger.Infof("AMQP message successfully published after automatic reconnection on attempt %d", attempt)
			}

			return nil
		}

		// Анализируем сбой рантайма драйвера через errors.As
		var linkErr *amqp.LinkError
		var sessionErr *amqp.SessionError
		var connErr *amqp.ConnError

		switch {
		case errors.As(err, &linkErr) || errors.As(err, &sessionErr) || errors.As(err, &connErr):
			// Поймали сетевой сбой (или сработавший idle timeout брокера) [INDEX].
			if attempt < maxAttempts {
				cs.logger.Warnf("AMQP network failure detected on attempt %d (%v). Invalidating resources and retrying...", attempt, err)
				cs.handleSendError(targetName, err)
				continue // Уходим на вторую попытку со свежим сокетом!
			}

			return errs.NewCommonError("azure sender network error persisted after retry", err)
		default:
			// Обычный таймаут контекста (context deadline exceeded) диспетчера.
			// Сеть исправна, реконнект не нужен — выходим сразу.
			return errs.NewCommonError("azure sender unrecoverable send error", err)
		}
	}

	return errs.NewCommonError("azure sender unexpected retry loop exit", nil)
}

//goland:noinspection DuplicatedCode
func (cs *ClientSender) establishConnection(ctx context.Context) error {
	var conn *amqp.Conn
	var err error

	// Ограничиваем контекст подключения на базе нашего ConnectTimeout из конфига!
	dialCtx, cancel := context.WithTimeout(ctx, cs.opts.connectTimeout)
	defer cancel()

	// Если в тесте передан мок — вызываем его, иначе идем в реальную сеть
	if cs.opts.dialFnTestGap != nil {
		conn, err = cs.opts.dialFnTestGap(dialCtx, cs.url, cs.opts.ConnOptions)
	} else {
		conn, err = amqp.Dial(dialCtx, cs.url, cs.opts.ConnOptions)
	}

	if err != nil {
		return errs.NewCommonError("dial failed", err)
	}

	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		_ = conn.Close()
		return errs.NewCommonError("failed to open session", err)
	}

	cs.conn = conn
	cs.session = session

	return nil
}

func (cs *ClientSender) getOrCreateSender(ctx context.Context, targetName string) (amqpSenderLink, error) {
	cs.mu.RLock()
	sender, exists := cs.senders[targetName]
	cs.mu.RUnlock()

	if exists && sender != nil {
		return sender, nil
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Double-checked locking
	if sender, exists = cs.senders[targetName]; exists && sender != nil {
		return sender, nil
	}

	if cs.session == nil || cs.conn == nil {
		if err := cs.establishConnection(ctx); err != nil {
			return nil, err
		}
	}

	cs.logger.Debugf("opening new amqp target link for target name: %s", targetName)
	newSender, err := cs.session.NewSender(ctx, targetName, nil)
	if err != nil {
		cs.session = nil

		return nil, errs.NewCommonError(fmt.Sprintf("failed to create amqp link for target [%s]", targetName), err)
	}

	cs.senders[targetName] = newSender

	return newSender, nil
}

func (cs *ClientSender) handleSendError(targetName string, err error) {
	var linkErr *amqp.LinkError
	var sessionErr *amqp.SessionError
	var connErr *amqp.ConnError

	cs.mu.Lock()
	defer cs.mu.Unlock()

	switch {
	case errors.As(err, &linkErr):
		cs.logger.Errorf("AMQP Link dead for address %s: %v. Cleaning target link.", targetName, linkErr)
		if sender, exists := cs.senders[targetName]; exists {
			_ = sender.Close(context.Background())
			delete(cs.senders, targetName)
		}

	case errors.As(err, &sessionErr):
		cs.logger.Errorf("AMQP Session dead: %v. Invalidating session and all target links.", sessionErr)
		cs.clearAllLinks()
		cs.session = nil

	case errors.As(err, &connErr):
		cs.logger.Errorf("AMQP Socket dead (or idle timeout): %v. Resetting entire cluster connection.", connErr)
		cs.clearAllLinks()
		cs.session = nil
		cs.conn = nil
	}
}

func (cs *ClientSender) resetInfrastructure() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.clearAllLinks()
	cs.session = nil
	cs.conn = nil
}

func (cs *ClientSender) clearAllLinks() {
	for addr, sender := range cs.senders {
		_ = sender.Close(context.Background())
		delete(cs.senders, addr)
	}
}

func (cs *ClientSender) pStr(s string) *string {
	return &s
}
