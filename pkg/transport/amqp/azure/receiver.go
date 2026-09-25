package azure

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type Receiver struct {
	opts   *ReceiverOptions
	link   AMQPReceiverLink // Наш единственный фиксированный линк-получатель
	logger logger.Logger
	mu     sync.RWMutex
	initMu sync.Mutex // Защищает ленивую инициализацию линка от Thundering Herd
	// поля для сбора рантайм-метрики со стороны Azure
	totalMessagesCounter atomic.Uint64 // Потокобезопасный счетчик успешных сообщений
	totalErrorsCounter   atomic.Uint64 // Потокобезопасный счетчик сетевых сбоев/ошибок
	connectedAt          time.Time     // Таймштамп момента успешного открытия линка
}

// Привязываем структуру к итоговому интерфейсу пакета абстракций
var _ pkgamqp.Receiver = (*Receiver)(nil)
var _ AMQPReceiver = (*Receiver)(nil)

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

func (r *Receiver) Receive(ctx context.Context) (pkgamqp.Message, error) {
	return r.ReceiveWithOpts(ctx, r.getReceiveOpts())
}

func (r *Receiver) ReceiveWithOpts(ctx context.Context, receiveOpts *amqp.ReceiveOptions) (pkgamqp.Message, error) {
	// Получаем или лениво инициализируем линк очереди/топика
	receiverLink, err := r.getReceiver(ctx)
	if err != nil {
		return nil, errs.NewTlCommonError("Receive", "azure receiver failed to get link", err)
	}

	// опции
	opts := receiveOpts
	if utils.IsNil(opts) {
		opts = r.getReceiveOpts()
	}
	// Читаем сообщение из сокета (блокирующий вызов библиотеки Azure)
	azureMsg, err := receiverLink.Receive(ctx, opts)
	if err != nil {
		r.totalErrorsCounter.Add(1) // <-- ИНКРЕМЕНТ: Фиксируем ошибку сети
		r.handleReceiverFailure(err)
		return nil, errs.NewTlCommonError("Receive", "azure receiver incoming packet error", err)
	}

	r.totalMessagesCounter.Add(1) // <-- ИНКРЕМЕНТ: Успешно прочитали пакет

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

	resMsg := &Message{
		Payload:    finalPayload,
		Props:      make(map[string]any),
		TargetName: r.opts.TargetName, // Фиксируем точный топик-источник, так как ресивер теперь 1-к-1
	}

	if azureMsg.ApplicationProperties != nil {
		maps.Copy(resMsg.Props, azureMsg.ApplicationProperties)
	}

	resMsg.Props[sysMsgKey] = azureMsg

	return resMsg, nil
}

func (r *Receiver) Accept(ctx context.Context, msg pkgamqp.Message) error {
	azureMsg, err := ExtractOriginalMessage(msg)
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

func (r *Receiver) Reject(ctx context.Context, msg pkgamqp.Message, err error) error {
	azureMsg, extractErr := ExtractOriginalMessage(msg)
	if extractErr != nil {
		return errs.NewTlCommonError("Reject", "extract original azure amqp message failed", extractErr)
	}

	receiverLink, localErr := r.getReceiver(ctx)
	if localErr != nil {
		return errs.NewTlCommonError("Reject", "retrieve azure receiver failed", localErr)
	}

	amqpErr := &amqp.Error{Condition: "amqp:processing-error", Description: err.Error()}
	if localErr = receiverLink.RejectMessage(ctx, azureMsg, amqpErr); localErr != nil {
		return errs.NewTlCommonError("Reject", "azure receiver failed to reject amqp message", localErr)
	}

	return nil
}

func (r *Receiver) Release(ctx context.Context, msg pkgamqp.Message) error {
	azureMsg, err := ExtractOriginalMessage(msg)
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

//goland:noinspection DuplicatedCode
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

func (r *Receiver) GetTargetName() string {
	return r.opts.TargetName
}

// Stats возвращает строго типизированный снимок состояния рантайма Azure AMQP (Service Bus).
// Заполняет блоки COMMON и AMQP, оставляя поля KAFKA пустыми (они скроются в JSON через omitempty).
func (r *Receiver) Stats() pkgamqp.ReceiverStats {
	r.mu.RLock()
	// Если соединение или линк еще не инициализированы
	if utils.IsNil(r.link) {
		r.mu.RUnlock()
		return pkgamqp.ReceiverStats{
			BrokerType: "azure-amqp",
			TargetName: r.opts.TargetName, // Имя Queue или Subscription
			Status:     "disconnected",
		}
	}

	// Извлекаем низкоуровневые метрики линка, если ваша обертка над azure/go-amqp их трекает.
	// В AMQP 1.0 статус проверяется через контекст или состояние линка r.link.Closed()
	status := "connected"
	// if r.link.Closed() { status = "disconnected" }

	// Считываем внутренние атомарные счетчики, которые ресивер обновляет при каждом успешном Receive() и ошибках
	totalMessages := r.totalMessagesCounter.Load()
	totalErrors := r.totalErrorsCounter.Load()

	r.mu.RUnlock()

	return pkgamqp.ReceiverStats{
		// ====================================================================
		// COMMON Specific
		// ====================================================================
		BrokerType:    "azure-amqp",
		TargetName:    r.opts.TargetName,
		Status:        status,
		TotalMessages: totalMessages, // Счетчик успешно обработанных пакетов фреймворком
		TotalErrors:   totalErrors,   // Счетчик сетевых сбоев/ошибок десериализации
		Lag:           -1,            // В AMQP 1.0 клиент не знает лаг. Для точного лага нужен Azure Management API (REST/SDK)
		ConnectedAt:   r.connectedAt,

		// ====================================================================
		// AMQP Specific
		// ====================================================================
		// PrefetchCount показывает, сколько сообщений брокер может отправить в буфер линка без подтверждения (Credits)
		PrefetchCount: int64(r.opts.LinkCredit),

		// Если архитектура позволяет узнать количество конкурирующих консьюмеров (опционально)
		ConsumerCount: 1,
	}
}

//goland:noinspection DuplicatedCode
func (r *Receiver) getReceiver(ctx context.Context) (AMQPReceiverLink, error) {
	// 1. Быстрый путь (Fast Path): если линк жив, отдаем под RLock за наносекунды
	r.mu.RLock()
	if !utils.IsNil(r.link) {
		r.mu.RUnlock()

		return r.link, nil
	}
	r.mu.RUnlock()

	// 2. Медленный путь (Slow Path) под мьютексом инициализации линка
	r.initMu.Lock()
	defer r.initMu.Unlock()

	// 3. Double-check
	r.mu.RLock()
	if !utils.IsNil(r.link) {
		r.mu.RUnlock()

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
	linkOpts := r.buildReceiverOpts()

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
	r.connectedAt = time.Now()
	r.mu.Unlock()

	return newReceiver, nil
}

func (r *Receiver) buildReceiverOpts() *amqp.ReceiverOptions {
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

	return linkOpts
}

func (r *Receiver) handleReceiverFailure(err error) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		r.logger.Warnf("AMQP timed out (DeadlineExceeded), do nothing, pass")

		return
	}

	r.logger.Warnf("AMQP packet reading failure detected: %v. Notifying connector...", err)

	// ИСПРАВЛЕНИЕ: Просто передаем сетевую ошибку в общий коннектор.
	// Он сам атомарно разберется, какой уровень инвалидировать (Session или Connection)
	r.opts.Connector.Invalidate(err)

	r.mu.Lock()
	r.link = nil // Сбрасываем локальный линк, чтобы на следующем Receive() лениво его пересоздать
	r.mu.Unlock()
}

func (r *Receiver) getReceiveOpts() *amqp.ReceiveOptions {
	if utils.IsNil(r.opts.ReceiveOpts) {
		return r.buildDefaultReceiveOpts()
	}

	return r.opts.ReceiveOpts
}

func (r *Receiver) buildDefaultReceiveOpts() *amqp.ReceiveOptions {
	return &amqp.ReceiveOptions{}
}
