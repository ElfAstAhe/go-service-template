package azure

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
)

// amqpReceiverLink описывает методы встроенного получателя Azure AMQP,
// необходимые для чтения, подтверждения и закрытия линка.
type amqpReceiverLink interface {
	Receive(ctx context.Context, opts *amqp.ReceiveOptions) (*amqp.Message, error)
	AcceptMessage(ctx context.Context, msg *amqp.Message) error
	RejectMessage(ctx context.Context, msg *amqp.Message, err *amqp.Error) error
	ReleaseMessage(ctx context.Context, msg *amqp.Message) error
	Close(ctx context.Context) error
}

// Ключ для скрытой передачи системного сообщения внутри pkgamqp.Message
const sysMsgKey = "_sys_amqp_orig_azure_message"

type ClientReceiver struct {
	mu        sync.RWMutex
	url       string
	logger    logger.Logger
	conn      *amqp.Conn
	session   *amqp.Session
	receivers map[string]amqpReceiverLink
	opts      *options
}

var _ pkgamqp.ClientReceiver = (*ClientReceiver)(nil)

func NewClientReceiver(url string, log logger.Logger, opts ...Option) *ClientReceiver {
	conf := &options{
		ConnOptions:     &amqp.ConnOptions{},
		shutdownTimeout: 3, // Безопасный дефолт
	}

	for _, opt := range opts {
		opt(conf)
	}

	return &ClientReceiver{
		url:       url,
		logger:    log.GetLogger("azure-client-receiver"),
		receivers: make(map[string]amqpReceiverLink),
		opts:      conf,
	}
}

//goland:noinspection DuplicatedCode
func (cr *ClientReceiver) Close(ctx context.Context) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cr.opts.shutdownTimeout)
	defer cancel()

	for queue, receiver := range cr.receivers {
		_ = receiver.Close(ctx)
		delete(cr.receivers, queue)
	}

	if cr.session != nil {
		_ = cr.session.Close(ctx)
		cr.session = nil
	}

	if cr.conn != nil {
		_ = cr.conn.Close()
		cr.conn = nil
	}

	return nil
}

// Receive блокирует поток, ожидая новое сообщение из указанной очереди брокера
// Receive блокирует поток, ожидая новое сообщение из указанной очереди брокера
func (cr *ClientReceiver) Receive(ctx context.Context, queueName string) (*pkgamqp.Message, error) {
	receiver, err := cr.getOrCreateReceiver(ctx, queueName)
	if err != nil {
		return nil, fmt.Errorf("azure receiver failed to get link for %s: %w", queueName, err)
	}

	// Читаем сообщение из сокета (блокирующий вызов библиотеки Azure)
	azureMsg, err := receiver.Receive(ctx, nil)
	if err != nil {
		cr.handleReceiverFailure(queueName, err)
		return nil, fmt.Errorf("azure receiver incoming packet error: %w", err)
	}

	// 1. Безопасно собираем Payload из Data [][]byte
	var finalPayload []byte
	if len(azureMsg.Data) > 0 {
		// Подсчитываем общий размер всех чанков для эффективного выделения памяти без лишних аллокаций
		totalSize := 0
		for _, chunk := range azureMsg.Data {
			totalSize += len(chunk)
		}

		finalPayload = make([]byte, 0, totalSize)
		for _, chunk := range azureMsg.Data {
			finalPayload = append(finalPayload, chunk...)
		}
	} else if azureMsg.Value != nil {
		// Фолбек-страховка: если отправитель упаковал данные как AMQP Value (например, чистую строку или байты)
		if byteVal, ok := azureMsg.Value.([]byte); ok {
			finalPayload = byteVal
		} else if strVal, ok := azureMsg.Value.(string); ok {
			finalPayload = []byte(strVal)
		}
	}

	// Мапим системное сообщение в наш чистый pkgamqp.Message
	resMsg := &pkgamqp.Message{
		Payload:    finalPayload,
		Properties: make(map[string]any),
	}

	// Переносим пользовательские ApplicationProperties, если они есть
	if azureMsg.ApplicationProperties != nil {
		for k, v := range azureMsg.ApplicationProperties {
			resMsg.Properties[k] = v
		}
	}

	// Скрытый хак: Прячем оригинальный *amqp.Message внутрь Properties,
	// чтобы методы Accept/Reject ниже знали, кого именно подтверждать брокеру.
	resMsg.Properties[sysMsgKey] = azureMsg

	return resMsg, nil
}

// Accept подтверждает успешную обработку сообщения
func (cr *ClientReceiver) Accept(ctx context.Context, msg *pkgamqp.Message) error {
	azureMsg, err := cr.extractSysMessage(msg)
	if err != nil {
		return err
	}

	receiver, err := cr.getOrCreateReceiver(ctx, "") // Поиск по сессии не требует имени
	if err != nil {
		return err
	}

	if err := receiver.AcceptMessage(ctx, azureMsg); err != nil {
		return fmt.Errorf("azure receiver failed to accept amqp message: %w", err)
	}
	return nil
}

// Reject уводит сообщение в Dead Letter Address (DLA) брокера при критической ошибке
func (cr *ClientReceiver) Reject(ctx context.Context, msg *pkgamqp.Message, err error) error {
	azureMsg, extractErr := cr.extractSysMessage(msg)
	if extractErr != nil {
		return extractErr
	}

	receiver, getErr := cr.getOrCreateReceiver(ctx, "")
	if getErr != nil {
		return getErr
	}

	// Передаем ошибку как причину отклонения пакета
	amqpErr := &amqp.Error{Condition: "amqp:processing-error", Description: err.Error()}
	if err := receiver.RejectMessage(ctx, azureMsg, amqpErr); err != nil {
		return fmt.Errorf("azure receiver failed to reject amqp message: %w", err)
	}
	return nil
}

// Release возвращает сообщение обратно в начало очереди для ретрая
func (cr *ClientReceiver) Release(ctx context.Context, msg *pkgamqp.Message) error {
	azureMsg, err := cr.extractSysMessage(msg)
	if err != nil {
		return err
	}

	receiver, err := cr.getOrCreateReceiver(ctx, "")
	if err != nil {
		return err
	}

	if err := receiver.ReleaseMessage(ctx, azureMsg); err != nil {
		return fmt.Errorf("azure receiver failed to release amqp message: %w", err)
	}
	return nil
}

// ---- Внутренние служебные методы (Оркестрация сокета) ----

func (cr *ClientReceiver) establishConnection(ctx context.Context) error {
	var conn *amqp.Conn
	var err error

	if cr.opts.dialFnTestGap != nil {
		conn, err = cr.opts.dialFnTestGap(ctx, cr.url, cr.opts.ConnOptions)
	} else {
		conn, err = amqp.Dial(ctx, cr.url, cr.opts.ConnOptions)
	}

	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to open session: %w", err)
	}

	cr.conn = conn
	cr.session = session
	return nil
}

func (cr *ClientReceiver) getOrCreateReceiver(ctx context.Context, queueName string) (amqpReceiverLink, error) {
	cr.mu.RLock()
	// Если queueName пустой (вызов из Accept/Reject), берем любой первый живой ресивер сессии
	if queueName == "" {
		for _, r := range cr.receivers {
			if r != nil {
				cr.mu.RUnlock()
				return r, nil
			}
		}
	}
	receiver, exists := cr.receivers[queueName]
	cr.mu.RUnlock()

	if exists && receiver != nil {
		return receiver, nil
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	// Double-checked locking
	if queueName != "" {
		if receiver, exists = cr.receivers[queueName]; exists && receiver != nil {
			return receiver, nil
		}
	}

	if cr.session == nil || cr.conn == nil {
		if err := cr.establishConnection(ctx); err != nil {
			return nil, err
		}
	}

	// Если имя пустое и сессия пустая — мы не можем создать дефолтный линк
	if queueName == "" {
		return nil, errors.New("amqp session is empty, cannot manage message acknowledgement state")
	}

	cr.logger.Debugf("opening new amqp source link for queue: %s", queueName)
	newReceiver, err := cr.session.NewReceiver(ctx, queueName, nil)
	if err != nil {
		cr.session = nil
		return nil, fmt.Errorf("failed to create amqp link receiver: %w", err)
	}

	cr.receivers[queueName] = newReceiver
	return newReceiver, nil
}

func (cr *ClientReceiver) handleReceiverFailure(queueName string, err error) {
	var linkErr *amqp.LinkError
	var sessionErr *amqp.SessionError
	var connErr *amqp.ConnError

	cr.mu.Lock()
	defer cr.mu.Unlock()

	switch {
	case errors.As(err, &linkErr):
		cr.logger.Errorf("AMQP Receiver Link dead for queue %s: %v. Cleaning target link.", queueName, linkErr)
		if receiver, exists := cr.receivers[queueName]; exists {
			_ = receiver.Close(context.Background())
			delete(cr.receivers, queueName)
		}

	case errors.As(err, &sessionErr):
		cr.logger.Errorf("AMQP Receiver Session dead: %v. Invalidating session and links.", sessionErr)
		cr.clearAllLinks()
		cr.session = nil

	case errors.As(err, &connErr):
		cr.logger.Errorf("AMQP Receiver Socket dead: %v. Resetting server connection pipelines.", connErr)
		cr.clearAllLinks()
		cr.session = nil
		cr.conn = nil
	}
}

func (cr *ClientReceiver) clearAllLinks() {
	for queue, receiver := range cr.receivers {
		_ = receiver.Close(context.Background())
		delete(cr.receivers, queue)
	}
}

func (cr *ClientReceiver) extractSysMessage(msg *pkgamqp.Message) (*amqp.Message, error) {
	if msg == nil || msg.Properties == nil {
		return nil, errors.New("cannot manage ack state for empty message envelope")
	}

	raw, exists := msg.Properties[sysMsgKey]
	if !exists {
		return nil, errors.New("message envelope missing underlying system amqp context")
	}

	sysMsg, ok := raw.(*amqp.Message)
	if !ok {
		return nil, errors.New("invalid underlying amqp packet structure type")
	}

	return sysMsg, nil
}
