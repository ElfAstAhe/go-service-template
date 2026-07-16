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

type Receiver struct {
	opts   *ReceiverOptions
	link   AmqpReceiverLink // Наш единственный фиксированный линк-получатель
	logger logger.Logger
	mu     sync.RWMutex
	initMu sync.Mutex // Защищает ленивую инициализацию линка от Thundering Herd
}

// Привязываем структуру к итоговому интерфейсу пакета абстракций
var _ pkgamqp.Receiver[*amqp.ReceiveOptions, *amqp.MessageHeader] = (*Receiver)(nil)

func NewReceiver(opts ...ReceiverOption) (*Receiver, error) {
	clientOpts := NewReceiverOptions() // Все базовые дефолты таймаутов и кредитов внутри

	for _, opt := range opts {
		opt(clientOpts)
	}

	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewReceiver", "client receiver options validate failed", err)
	}

	return &Receiver{
		opts:   clientOpts,
		logger: clientOpts.Logger.GetLogger("azure-amqp-receiver"),
	}, nil
}

func (r *Receiver) Receive(ctx context.Context, receiveOpts *amqp.ReceiveOptions) (*pkgamqp.Message[*amqp.MessageHeader], error) {
	// Получаем или лениво инициализируем линк очереди/топика
	receiverLink, err := r.getReceiver(ctx)
	if err != nil {
		return nil, errs.NewTlCommonError("Receive", "azure receiver failed to get link", err)
	}

	// Читаем сообщение из сокета (блокирующий вызов библиотеки Azure)
	azureMsg, err := receiverLink.Receive(ctx, receiveOpts)
	if err != nil {
		r.handleReceiverFailure(err)
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
		TargetName: r.opts.TargetName, // Фиксируем точный топик-источник, так как ресивер теперь 1-к-1
	}

	if azureMsg.ApplicationProperties != nil {
		for k, v := range azureMsg.ApplicationProperties {
			resMsg.Properties[k] = v
		}
	}

	resMsg.Properties[sysMsgKey] = azureMsg

	return resMsg, nil
}

func (r *Receiver) Accept(ctx context.Context, msg *pkgamqp.Message[*amqp.MessageHeader]) error {
	azureMsg, err := r.extractOriginalMessage(msg)
	if err != nil {
		return errs.NewTlCommonError("Accept", "extract original azure amqp message failed", err)
	}

	receiverLink, err := r.getReceiver(ctx)
	if err != nil {
		return errs.NewTlCommonError("Accept", "retrieve azure receiver failed", err)
	}

	if err = receiverLink.AcceptMessage(ctx, azureMsg); err != nil {
		return errs.NewTlCommonError("Accept", "azure receiver failed to accept amqp message", err)
	}

	return nil
}

func (r *Receiver) Reject(ctx context.Context, msg *pkgamqp.Message[*amqp.MessageHeader], err error) error {
	azureMsg, extractErr := r.extractOriginalMessage(msg)
	if extractErr != nil {
		return errs.NewTlCommonError("Reject", "extract original azure amqp message failed", extractErr)
	}

	receiverLink, err := r.getReceiver(ctx)
	if err != nil {
		return errs.NewTlCommonError("Reject", "retrieve azure receiver failed", err)
	}

	amqpErr := &amqp.Error{Condition: "amqp:processing-error", Description: err.Error()}
	if err = receiverLink.RejectMessage(ctx, azureMsg, amqpErr); err != nil {
		return errs.NewTlCommonError("Reject", "azure receiver failed to reject amqp message", err)
	}

	return nil
}

func (r *Receiver) Release(ctx context.Context, msg *pkgamqp.Message[*amqp.MessageHeader]) error {
	azureMsg, err := r.extractOriginalMessage(msg)
	if err != nil {
		return errs.NewTlCommonError("Release", "extract original azure amqp message failed", err)
	}

	receiverLink, err := r.getReceiver(ctx)
	if err != nil {
		return errs.NewTlCommonError("Release", "retrieve azure receiver failed", err)
	}

	if err = receiverLink.ReleaseMessage(ctx, azureMsg); err != nil {
		return errs.NewTlCommonError("Release", "azure receiver failed to release amqp message", err)
	}

	return nil
}

func (r *Receiver) Close(ctx context.Context) error {
	r.logger.Debug("close started")
	defer r.logger.Debug("close finished")

	r.mu.Lock()
	linkToClose := r.link
	r.link = nil // Мгновенно обнуляем ссылку
	r.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, r.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)

	go func() {
		var closeErrs []error
		// Мягко гасим исключительно СВОЙ линк-получатель.
		// Сессию и коннект не трогаем — за них отвечает глобальный Connector.
		if !utils.IsNil(linkToClose) {
			if err := linkToClose.Close(closeCtx); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}
		done <- errors.Join(closeErrs...)
	}()

	select {
	case err := <-done:
		if err != nil {
			return errs.NewTlCommonError("Close", "Azure AMQP client receiver close fails", err)
		}
		return nil
	case <-closeCtx.Done():
		return errs.NewTlCommonError("Close", "Azure AMQP client receiver close timeout", closeCtx.Err())
	}
}

func (r *Receiver) getReceiver(ctx context.Context) (AmqpReceiverLink, error) {
	// 1. Быстрый путь (Fast Path): если линк жив, отдаем под RLock за наносекунды
	r.mu.RLock()
	if !utils.IsNil(r.link) {
		defer r.mu.RUnlock()
		return r.link, nil
	}
	r.mu.RUnlock()

	// 2. Медленный путь (Slow Path) под мьютексом инициализации линка
	r.initMu.Lock()
	defer r.initMu.Unlock()

	// 3. Double-check
	r.mu.RLock()
	if !utils.IsNil(r.link) {
		defer r.mu.RUnlock()
		return r.link, nil
	}
	r.mu.RUnlock()

	// 4. Запрашиваем живую общую сессию у внешнего коннектора
	session, err := r.opts.Connector.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	r.logger.Debugf("opening new durable amqp source link for target name: %s", r.opts.TargetName)

	linkCtx, cancel := context.WithTimeout(ctx, r.opts.ConnectTimeout)
	defer cancel()

	// Настраиваем Durable-подписку и динамические кредиты из файла настроек
	linkOpts := &amqp.ReceiverOptions{
		Durability:   amqp.DurabilityConfiguration,
		ExpiryPolicy: amqp.ExpiryPolicyNever,
		Credit:       r.opts.LinkCredit, // Твоя инкапсулированная настройка int32
	}

	// Если пользователь передал целиком кастомные ReceiverOpts через WithReceiverOpts,
	// берем их, но страхуем поле Credit, если оно там не заполнено (0)
	if r.opts.ReceiverOpts != nil {
		linkOpts = r.opts.ReceiverOpts
		if linkOpts.Credit == 0 {
			linkOpts.Credit = r.opts.LinkCredit
		}
	}

	newReceiver, err := session.NewReceiver(linkCtx, r.opts.TargetName, linkOpts)
	if err != nil {
		// Ошибка создания линка — сообщаем коннектору, сбрасывая локальное состояние
		r.opts.Connector.Invalidate(err)
		r.mu.Lock()
		r.link = nil
		r.mu.Unlock()
		return nil, errs.NewTlCommonError("getReceiver", fmt.Sprintf("failed to create amqp receiver link for [%s]", r.opts.TargetName), err)
	}

	r.mu.Lock()
	r.link = newReceiver
	r.mu.Unlock()

	return newReceiver, nil
}

func (r *Receiver) handleReceiverFailure(err error) {
	r.logger.Warnf("AMQP packet reading failure detected: %v. Notifying connector...", err)

	// ИСПРАВЛЕНИЕ: Просто передаем сетевую ошибку в общий коннектор.
	// Он сам атомарно разберется, какой уровень инвалидировать (Session или Connection)
	r.opts.Connector.Invalidate(err)

	r.mu.Lock()
	r.link = nil // Сбрасываем локальный линк, чтобы на следующем Receive() лениво его пересоздать
	r.mu.Unlock()
}

func (r *Receiver) extractOriginalMessage(msg *pkgamqp.Message[*amqp.MessageHeader]) (*amqp.Message, error) {
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

func (r *Receiver) GetTargetName() string {
	return r.opts.TargetName
}
