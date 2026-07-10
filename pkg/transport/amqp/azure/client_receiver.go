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

const sysMsgKey = "_sys_amqp_orig_azure_message"

type ClientReceiver struct {
	mu         sync.RWMutex
	initMu     sync.Mutex // Изолирует Thundering Herd при сетевом Dial/Link Open
	url        string
	logger     logger.Logger
	connection *amqp.Conn
	session    *amqp.Session
	receivers  map[string]AmqpReceiverLink
	opts       *ClientReceiverOptions
}

var _ pkgamqp.ClientReceiver[*amqp.ReceiveOptions, *amqp.MessageHeader] = (*ClientReceiver)(nil)

func NewClientReceiver(opts ...ReceiverOption) (*ClientReceiver, error) {
	clientOpts := NewClientReceiverOptions() // Все дефолты ( timeouts, delays) уже внутри

	for _, opt := range opts {
		opt(clientOpts)
	}

	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewClientReceiver", "client receiver options validate failed", err)
	}

	return &ClientReceiver{
		url:       clientOpts.URL,
		logger:    clientOpts.Logger.GetLogger("azure-amqp-client-receiver"),
		receivers: make(map[string]AmqpReceiverLink),
		opts:      clientOpts,
	}, nil
}

func (cr *ClientReceiver) Receive(ctx context.Context, targetName string, receiveOpts *amqp.ReceiveOptions) (*pkgamqp.Message[*amqp.MessageHeader], error) {
	if targetName == "" {
		return nil, errs.NewTlCommonError("Receive", "target queue/topic name cannot be empty", nil)
	}

	receiver, err := cr.getOrCreateReceiver(ctx, targetName)
	if err != nil {
		return nil, errs.NewTlCommonError("Receive", fmt.Sprintf("azure receiver failed to get link for %s", targetName), err)
	}

	azureMsg, err := receiver.Receive(ctx, receiveOpts)
	if err != nil {
		cr.handleReceiverFailure(targetName, err)
		return nil, errs.NewTlCommonError("Receive", "azure receiver incoming packet error", err)
	}

	// Высокопроизводительная сборка Payload через copy без лишних микро-аллокаций в куче
	var finalPayload []byte
	if len(azureMsg.Data) > 0 {
		totalSize := 0
		for _, chunk := range azureMsg.Data {
			totalSize += len(chunk)
		}

		finalPayload = make([]byte, totalSize)
		offset := 0
		for _, chunk := range azureMsg.Data {
			offset += copy(finalPayload[offset:], chunk)
		}
	} else if azureMsg.Value != nil {
		if byteVal, ok := azureMsg.Value.([]byte); ok {
			finalPayload = byteVal
		} else if strVal, ok := azureMsg.Value.(string); ok {
			finalPayload = []byte(strVal)
		}
	}

	resMsg := &pkgamqp.Message[*amqp.MessageHeader]{
		Payload:    finalPayload,
		Properties: make(map[string]any),
		TargetName: targetName, // Фиксируем точный топик-источник для адресного ACK
	}

	if azureMsg.ApplicationProperties != nil {
		for k, v := range azureMsg.ApplicationProperties {
			resMsg.Properties[k] = v
		}
	}

	resMsg.Properties[sysMsgKey] = azureMsg

	return resMsg, nil
}

func (cr *ClientReceiver) Accept(ctx context.Context, msg *pkgamqp.Message[*amqp.MessageHeader]) error {
	azureMsg, err := cr.extractOriginalMessage(msg)
	if err != nil {
		return errs.NewTlCommonError("Accept", "extract original azure amqp message failed", err)
	}

	// ИСПРАВЛЕНИЕ: Ищем ресивер строго по имени SourceQueue, откуда сообщение пришло
	receiver, err := cr.getOrCreateReceiver(ctx, msg.TargetName)
	if err != nil {
		return errs.NewTlCommonError("Accept", "retrieve or create azure receiver failed", err)
	}

	if err = receiver.AcceptMessage(ctx, azureMsg); err != nil {
		return errs.NewTlCommonError("Accept", "azure receiver failed to accept amqp message", err)
	}

	return nil
}

func (cr *ClientReceiver) Reject(ctx context.Context, msg *pkgamqp.Message[*amqp.MessageHeader], err error) error {
	azureMsg, extractErr := cr.extractOriginalMessage(msg)
	if extractErr != nil {
		return errs.NewTlCommonError("Reject", "extract original azure amqp message failed", extractErr)
	}

	receiver, getErr := cr.getOrCreateReceiver(ctx, msg.TargetName)
	if getErr != nil {
		return errs.NewTlCommonError("Reject", "retrieve or create azure receiver failed", getErr)
	}

	amqpErr := &amqp.Error{Condition: "amqp:processing-error", Description: err.Error()}
	if err = receiver.RejectMessage(ctx, azureMsg, amqpErr); err != nil {
		return errs.NewTlCommonError("Reject", "azure receiver failed to reject amqp message", err)
	}

	return nil
}

func (cr *ClientReceiver) Release(ctx context.Context, msg *pkgamqp.Message[*amqp.MessageHeader]) error {
	azureMsg, err := cr.extractOriginalMessage(msg)
	if err != nil {
		return errs.NewTlCommonError("Release", "extract original azure amqp message failed", err)
	}

	receiver, err := cr.getOrCreateReceiver(ctx, msg.TargetName)
	if err != nil {
		return errs.NewTlCommonError("Release", "retrieve or create azure receiver failed", err)
	}

	if err = receiver.ReleaseMessage(ctx, azureMsg); err != nil {
		return errs.NewTlCommonError("Release", "azure receiver failed to release amqp message", err)
	}

	return nil
}

//goland:noinspection DuplicatedCode
func (cr *ClientReceiver) Close(ctx context.Context) error {
	cr.logger.Debugf("close started")
	defer cr.logger.Debugf("close finished")

	cr.mu.Lock()
	receiversToClose := make(map[string]AmqpReceiverLink, len(cr.receivers))
	for k, v := range cr.receivers {
		receiversToClose[k] = v
	}
	cr.receivers = make(map[string]AmqpReceiverLink)

	sessionToClose := cr.session
	connToClose := cr.connection

	cr.session = nil
	cr.connection = nil
	cr.mu.Unlock() // МЬЮТЕКС МОМЕНТАЛЬНО СВОБОДЕН

	closeCtx, closeCancel := context.WithTimeout(ctx, cr.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)

	go func() {
		var closeErrs = utils.NewConcurrentList[error]()
		var wg sync.WaitGroup

		for _, receiver := range receiversToClose {
			if utils.IsNil(receiver) {
				continue
			}
			wg.Add(1)
			go func(r AmqpReceiverLink) {
				defer wg.Done()
				if err := r.Close(closeCtx); err != nil {
					closeErrs.Append(err)
				}
			}(receiver)
		}
		wg.Wait()

		if !utils.IsNil(sessionToClose) {
			if err := sessionToClose.Close(closeCtx); err != nil {
				closeErrs.Append(err)
			}
		}

		if !utils.IsNil(connToClose) {
			if err := connToClose.Close(); err != nil {
				closeErrs.Append(err)
			}
		}

		done <- errors.Join(closeErrs.Snapshot()...)
	}()

	select {
	case err := <-done:
		if err != nil {
			return errs.NewTlCommonError("Close", "Azure AMQP client receiver close fails", err)
		}

		return nil
	case <-closeCtx.Done():
		if connToClose != nil {
			_ = connToClose.Close()
		}

		return errs.NewTlCommonError("Close", "Azure AMQP client receiver close timeout", closeCtx.Err())
	}
}

func (cr *ClientReceiver) GetTargetNames() []string {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var targetNames = make([]string, 0, len(cr.receivers))
	for key := range cr.receivers {
		targetNames = append(targetNames, key)
	}

	return targetNames
}

//goland:noinspection DuplicatedCode
func (cr *ClientReceiver) getOrCreateReceiver(ctx context.Context, targetName string) (AmqpReceiverLink, error) {
	if targetName == "" {
		return nil, errs.NewTlCommonError("getOrCreateReceiver", "target queue name empty", nil)
	}

	// 1. Fast Path под RLock
	cr.mu.RLock()
	receiver, exists := cr.receivers[targetName]
	cr.mu.RUnlock()

	if exists && !utils.IsNil(receiver) {
		return receiver, nil
	}

	// 2. Slow Path под мьютексом инициализации пула
	cr.initMu.Lock()
	defer cr.initMu.Unlock()

	// 3. Double Check
	cr.mu.RLock()
	receiver, exists = cr.receivers[targetName]
	cr.mu.RUnlock()

	if exists && !utils.IsNil(receiver) {
		return receiver, nil
	}

	session, err := cr.getOrCreateSession(ctx)
	if err != nil {
		return nil, err
	}

	cr.logger.Debugf("opening new durable amqp source link for target name: %s", targetName)

	linkCtx, cancel := context.WithTimeout(ctx, cr.opts.ConnectTimeout)
	defer cancel()

	// ИСПРАВЛЕНИЕ: Настраиваем Durable-подписку, чтобы Artemis хранил логи на диске во время деплоев
	linkOpts := &amqp.ReceiverOptions{
		Durability:   amqp.DurabilityConfiguration,
		ExpiryPolicy: amqp.ExpiryPolicyNever,
		Credit:       (int32)(DefaultReceiverCredit),
	}

	newReceiver, err := session.NewReceiver(linkCtx, targetName, linkOpts)
	if err != nil {
		cr.mu.Lock()
		cr.session = nil
		delete(cr.receivers, targetName)
		cr.mu.Unlock()
		return nil, errs.NewTlCommonError("getOrCreateReceiver", fmt.Sprintf("failed to create amqp receiver link for [%s]", targetName), err)
	}

	cr.mu.Lock()
	cr.receivers[targetName] = newReceiver
	cr.mu.Unlock()

	return newReceiver, nil
}

//goland:noinspection DuplicatedCode
func (cr *ClientReceiver) getOrCreateSession(ctx context.Context) (*amqp.Session, error) {
	cr.mu.RLock()
	connAlive := !utils.IsNil(cr.connection)
	sessAlive := !utils.IsNil(cr.session)
	localConn := cr.connection
	localSess := cr.session
	cr.mu.RUnlock()

	if connAlive && sessAlive {
		return localSess, nil
	}

	var newlyDialed bool

	if !connAlive {
		dialCtx, cancel := context.WithTimeout(ctx, cr.opts.ConnectTimeout)
		defer cancel()

		var conn *amqp.Conn
		var err error
		if !utils.IsNil(cr.opts.DialFnTestGap) {
			conn, err = cr.opts.DialFnTestGap(dialCtx, cr.url, cr.opts.ConnOpts)
		} else {
			conn, err = amqp.Dial(dialCtx, cr.url, cr.opts.ConnOpts)
		}
		if err != nil {
			return nil, errs.NewTlCommonError("getOrCreateSession", "dial failed", err)
		}
		localConn = conn
		newlyDialed = true
	}

	sessCtx, cancelSess := context.WithTimeout(ctx, cr.opts.ConnectTimeout)
	defer cancelSess()

	session, err := localConn.NewSession(sessCtx, cr.opts.SessionOpts)
	if err != nil {
		cr.mu.Lock()
		var connToCloseBehindMutex *amqp.Conn

		if newlyDialed {
			connToCloseBehindMutex = localConn
		} else {
			connToCloseBehindMutex = localConn
			cr.connection = nil
		}
		cr.session = nil

		cr.receivers = make(map[string]AmqpReceiverLink) // Инвалидируем весь кэш ресиверов
		cr.mu.Unlock()
		if connToCloseBehindMutex != nil {
			go func(c *amqp.Conn) {
				_ = c.Close() // Асинхронное закрытие сокета вне мьютекса
			}(connToCloseBehindMutex)
		}

		return nil, errs.NewTlCommonError("getOrCreateSession", "failed to open session", err)
	}
	cr.mu.Lock()
	cr.connection = localConn
	cr.session = session
	cr.mu.Unlock()

	return session, nil
}
func (cr *ClientReceiver) handleReceiverFailure(targetName string, err error) {
	var linkErr *amqp.LinkError
	var sessionErr *amqp.SessionError
	var connErr *amqp.ConnError

	cr.mu.Lock()
	defer cr.mu.Unlock()

	switch {
	case errors.As(err, &linkErr):
		cr.logger.Errorf("AMQP Receiver Link dead for target name %s: %v. Cleaning link.", targetName, linkErr)
		delete(cr.receivers, targetName)
		// Стираем только этот битый топик
	case errors.As(err, &sessionErr):
		cr.logger.Errorf("AMQP Receiver Session dead: %v. Invalidating session and pool.", sessionErr)
		cr.session = nil
		cr.receivers = make(map[string]AmqpReceiverLink)
	case errors.As(err, &connErr):
		cr.logger.Errorf("AMQP Receiver Socket dead: %v. Resetting full connector pipelines.", connErr)
		cr.connection = nil
		cr.session = nil
		cr.receivers = make(map[string]AmqpReceiverLink)
	}
}

func (cr *ClientReceiver) extractOriginalMessage(msg *pkgamqp.Message[*amqp.MessageHeader]) (*amqp.Message, error) {
	if msg == nil || msg.Properties == nil {
		return nil, errs.NewTlCommonError("extractOriginalMessage", "cannot manage ack state for empty envelope", nil)
	}
	raw, exists := msg.Properties[sysMsgKey]
	if !exists {
		return nil, errs.NewTlCommonError("extractOriginalMessage", "missing underlying system amqp context", nil)
	}
	sysMsg, ok := raw.(*amqp.Message)
	if !ok {
		return nil, errs.NewTlCommonError("extractOriginalMessage", "invalid underlying packet structure type", nil)
	}

	return sysMsg, nil
}
