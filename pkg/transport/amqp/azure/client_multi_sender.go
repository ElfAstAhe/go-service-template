package azure

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type ClientMultiSender struct {
	opts       *ClientSenderOptions
	connection *amqp.Conn
	session    *amqp.Session
	senders    map[string]AmqpSenderLink // Карта пула линков: targetName -> *amqp.Sender
	logger     logger.Logger
	mu         sync.RWMutex
	initMu     sync.Mutex // Защищает общие Connection/Session и создание новых линков
}

// Точное сопоставление дженерик-интерфейса мульти-сендера (все три типа — указатели, как в интерфейсе!)
var _ pkgamqp.ClientMultiSender[*amqp.SenderOptions, *amqp.SendOptions, *amqp.MessageHeader] = (*ClientMultiSender)(nil)

func NewClientMultiSender(opts ...SenderOption) (*ClientMultiSender, error) {
	clientOpts := NewClientSenderOptions() // Применяем базовые дефолты бэккоффа

	for _, opt := range opts {
		opt(clientOpts)
	}

	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewClientMultiSender", "client sender options validate failed", err)
	}

	return &ClientMultiSender{
		opts:    clientOpts,
		senders: make(map[string]AmqpSenderLink),
		logger:  clientOpts.Logger.GetLogger("azure-amqp-client-multi-sender"),
	}, nil
}

func (cms *ClientMultiSender) Publish(
	ctx context.Context,
	targetName string,
	senderOpts *amqp.SenderOptions, // Дженерик тип 1
	msg *pkgamqp.Message[*amqp.MessageHeader], // Дженерик тип 3
	sendOpts *amqp.SendOptions, // Дженерик тип 2
) error {
	cms.logger.Debugf("multi-publish to %s started", targetName)
	defer cms.logger.Debugf("multi-publish to %s finished", targetName)

	if targetName == "" {
		return errs.NewTlCommonError("Publish", "cannot publish to empty target name", nil)
	}
	if utils.IsNil(msg) {
		return errs.NewTlCommonError("Publish", "cannot publish nil message", nil)
	}

	for attempt := 1; attempt <= cms.opts.PublishMaxTryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return errs.NewTlCommonError("Publish", "context canceled before attempt", err)
		}

		cms.logger.Debugf("publish to %s, attempt #%d", targetName, attempt)

		// Получаем или лениво инициализируем линк для конкретной очереди/топика
		sender, err := cms.getOrCreateSender(ctx, targetName, senderOpts)
		if err != nil {
			if attempt < cms.opts.PublishMaxTryAttempts {
				cms.logger.Warnf("AMQP failed to get sender for %s on attempt %d: %v", targetName, attempt, err)
				cms.waitBackoff(ctx, attempt)
				continue
			}

			return errs.NewTlCommonError("Publish", fmt.Sprintf("azure multi-sender failed to initialize link for %s", targetName), err)
		}

		azureMsg := cms.prepareMessage(msg)

		err = sender.Send(ctx, azureMsg, sendOpts)
		if err == nil {
			cms.logger.Debugf("AMQP message successfully published to %s on attempt %d", targetName, attempt)
			return nil
		}

		// Обрабатываем сетевую ошибку отправки (сбрасывает ресурсы под мьютексом)
		err = cms.handleSendError(targetName, attempt, err)
		if err != nil {
			return err
		}

		cms.waitBackoff(ctx, attempt)
	}

	return errs.NewTlCommonError("Publish", "azure multi-sender unexpected retry loop exit", nil)
}

//goland:noinspection DuplicatedCode
func (cms *ClientMultiSender) Close(ctx context.Context) error {
	cms.logger.Debugf("close started")
	defer cms.logger.Debugf("close finished")

	cms.mu.Lock()
	// Копируем ссылки на карту линков, сессию и коннект для закрытия вне Lock
	sendersToClose := make(map[string]AmqpSenderLink, len(cms.senders))
	for k, v := range cms.senders {
		sendersToClose[k] = v
	}
	cms.senders = make(map[string]AmqpSenderLink)

	sessionToClose := cms.session
	connToClose := cms.connection

	cms.session = nil
	cms.connection = nil
	cms.mu.Unlock() // МЬЮТЕКС МГНОВЕННО СВОБОДЕН!

	closeCtx, closeCancel := context.WithTimeout(ctx, cms.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)

	go func() {
		var closeErrs []error
		var wg sync.WaitGroup
		var muErrs sync.Mutex

		// Мягко и параллельно закрываем все накопленные линки-сендеры топиков
		for _, sender := range sendersToClose {
			if utils.IsNil(sender) {
				continue
			}
			wg.Add(1)
			go func(s AmqpSenderLink) {
				defer wg.Done()
				if err := s.Close(closeCtx); err != nil {
					muErrs.Lock()
					closeErrs = append(closeErrs, err)
					muErrs.Unlock()
				}
			}(sender)
		}
		wg.Wait()

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
			return errs.NewTlCommonError("Close", "Azure AMQP client multi-sender close fails", err)
		}
		return nil
	case <-closeCtx.Done():
		if connToClose != nil {
			_ = connToClose.Close()
		}
		return errs.NewTlCommonError("Close", "Azure AMQP client multi-sender close timeout: connection force closed", closeCtx.Err())
	}
}

//lint:ignore U1000 Suppress unused warning if staticcheck complains during templates lifecycle
//goland:noinspection GoUnusedMethod
func (cms *ClientMultiSender) GetTargetNames() []string {
	cms.mu.RLock()
	defer cms.mu.RUnlock()

	// Честное копирование ключей под RLock для предотвращения Data Race в рантайме
	names := make([]string, 0, len(cms.senders))
	for name := range cms.senders {
		names = append(names, name)
	}
	return names
}

func (cms *ClientMultiSender) getOrCreateSender(ctx context.Context, targetName string, senderOpts *amqp.SenderOptions) (AmqpSenderLink, error) {
	// 1. Быстрый путь (Fast Path): если линк для топика жив, отдаем под RLock
	cms.mu.RLock()
	sender, exists := cms.senders[targetName]
	cms.mu.RUnlock()

	if exists && !utils.IsNil(sender) {
		return sender, nil
	}

	// 2. Медленный путь (Slow Path) под мьютексом инициализации пула
	cms.initMu.Lock()
	defer cms.initMu.Unlock()

	// 3. Double-check
	cms.mu.RLock()
	sender, exists = cms.senders[targetName]
	cms.mu.RUnlock()

	if exists && !utils.IsNil(sender) {
		return sender, nil
	}

	// 4. Проверяем и восстанавливаем общее соединение/сессию пула
	session, err := cms.getOrCreateSession(ctx)
	if err != nil {
		return nil, err
	}

	cms.logger.Debugf("opening new dynamic amqp target link for target name: %s", targetName)

	linkCtx, cancel := context.WithTimeout(ctx, cms.opts.ConnectTimeout)
	defer cancel()

	// Используем опции, если они переданы на вызов. Иначе берем дефолты.
	activeSenderOpts := cms.opts.SenderOpts
	if !utils.IsNil(senderOpts) {
		activeSenderOpts = senderOpts
	}

	newSender, err := session.NewSender(linkCtx, targetName, activeSenderOpts)
	if err != nil {
		cms.mu.Lock()
		cms.session = nil
		delete(cms.senders, targetName) // Сбрасываем только этот битый топик из карты
		cms.mu.Unlock()
		return nil, errs.NewTlCommonError("getOrCreateSender", fmt.Sprintf("failed to create amqp link for target [%s]", targetName), err)
	}

	// 5. Сохраняем готовый сендер в карту
	cms.mu.Lock()
	cms.senders[targetName] = newSender
	cms.mu.Unlock()

	return newSender, nil
}

func (cms *ClientMultiSender) getOrCreateSession(ctx context.Context) (*amqp.Session, error) {
	cms.mu.RLock()
	connAlive := !utils.IsNil(cms.connection)
	sessAlive := !utils.IsNil(cms.session)
	localConn := cms.connection
	localSess := cms.session
	cms.mu.RUnlock()

	if connAlive && sessAlive {
		return localSess, nil
	}

	var newlyDialed bool

	if !connAlive {
		dialCtx, cancel := context.WithTimeout(ctx, cms.opts.ConnectTimeout)
		defer cancel()

		var conn *amqp.Conn
		var err error
		if !utils.IsNil(cms.opts.DialFnTestGap) {
			conn, err = cms.opts.DialFnTestGap(dialCtx, cms.opts.URL, cms.opts.ConnOpts)
		} else {
			conn, err = amqp.Dial(dialCtx, cms.opts.URL, cms.opts.ConnOpts)
		}
		if err != nil {
			return nil, errs.NewTlCommonError("getOrCreateSession", "dial failed", err)
		}
		localConn = conn
		newlyDialed = true
	}

	sessCtx, cancelSess := context.WithTimeout(ctx, cms.opts.ConnectTimeout)
	defer cancelSess()

	session, err := localConn.NewSession(sessCtx, cms.opts.SessionOpts)
	if err != nil {
		// Асинхронное закрытие без удержания мьютекса
		cms.mu.Lock()
		var connToCloseBehindMutex *amqp.Conn

		if newlyDialed {
			connToCloseBehindMutex = localConn
		} else {
			connToCloseBehindMutex = localConn
			cms.connection = nil
		}
		cms.session = nil
		cms.senders = make(map[string]AmqpSenderLink) // Сбрасываем кэш, так как сессия упала
		cms.mu.Unlock()

		if connToCloseBehindMutex != nil {
			go func(c *amqp.Conn) {
				_ = c.Close()
			}(connToCloseBehindMutex)
		}

		return nil, errs.NewTlCommonError("getOrCreateSession", "failed to open session", err)
	}

	cms.mu.Lock()
	cms.connection = localConn
	cms.session = session
	cms.mu.Unlock()

	return session, nil
}

func (cms *ClientMultiSender) handleSendError(targetName string, attempt int, err error) error {
	var linkErr *amqp.LinkError
	var sessionErr *amqp.SessionError
	var connErr *amqp.ConnError

	cms.mu.Lock()
	defer cms.mu.Unlock()

	isNetworkErr := errors.As(err, &linkErr) || errors.As(err, &sessionErr) || errors.As(err, &connErr)

	if isNetworkErr {
		if attempt < cms.opts.PublishMaxTryAttempts {
			cms.logger.Warnf("AMQP network failure detected on publish to %s (attempt %d): %v. Invalidating resources...", targetName, attempt, err)

			switch {
			case errors.As(err, &linkErr):
				delete(cms.senders, targetName)
			case errors.As(err, &sessionErr):
				cms.session = nil
				cms.senders = make(map[string]AmqpSenderLink)
			case errors.As(err, &connErr):
				cms.connection = nil
				cms.session = nil
				cms.senders = make(map[string]AmqpSenderLink)
			}
			return nil
		}
		return errs.NewTlCommonError("Publish", fmt.Sprintf("azure multi-sender network error persisted for %s after retry %d", targetName, attempt), err)
	}
	return errs.NewTlCommonError("Publish", "azure multi-sender unrecoverable send error", err)
}

//goland:noinspection DuplicatedCode
func (cms *ClientMultiSender) waitBackoff(ctx context.Context, attempt int) {
	shift := uint(attempt - 1)
	if shift > 31 {
		shift = 31
	}
	delay := cms.opts.PublishBaseRetryDelay * (1 << shift)
	if delay > cms.opts.PublishMaxRetryDelay || delay <= 0 {
		delay = cms.opts.PublishMaxRetryDelay
	}
	delayMs := int(delay / time.Millisecond)
	if delayMs > 5 {
		maxJitterMs := delayMs / 5
		jitterMs := rand.IntN(maxJitterMs)
		jitter := time.Duration(jitterMs) * time.Millisecond
		if rand.IntN(2) == 0 {
			delay += jitter
		} else {
			delay -= jitter
		}
	}
	cms.logger.Debugf("backing off for %v before next attempt", delay)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (cms *ClientMultiSender) prepareMessage(msg *pkgamqp.Message[*amqp.MessageHeader]) *amqp.Message {
	azureMsg := amqp.NewMessage(msg.Payload)
	if !utils.IsNil(msg.Header) {
		azureMsg.Header = msg.Header
	}
	azureMsg.Properties = &amqp.MessageProperties{
		ContentType: cms.pStr("application/json"),
	}
	if len(msg.Properties) > 0 {
		azureMsg.ApplicationProperties = msg.Properties
	}
	return azureMsg
}

func (cms *ClientMultiSender) pStr(s string) *string {
	return &s
}
