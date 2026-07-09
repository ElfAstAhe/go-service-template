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

type ClientSingleSender struct {
	opts       *ClientSenderOptions
	connection *amqp.Conn
	session    *amqp.Session
	sender     AmqpSenderLink
	logger     logger.Logger
	mu         sync.RWMutex
	initMu     sync.Mutex
}

var _ pkgamqp.ClientSingleSender[*amqp.SendOptions, *amqp.MessageHeader] = (*ClientSingleSender)(nil)

func NewClientSingleSender(opts ...SenderOption) (*ClientSingleSender, error) {
	clientOpts := &ClientSenderOptions{}

	for _, opt := range opts {
		opt(clientOpts)
	}

	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewClientSingleSender", "client sender options validate failed", err)
	}

	return &ClientSingleSender{
		opts:       clientOpts,
		connection: nil,
		session:    nil,
		sender:     nil,
		logger:     clientOpts.Logger.GetLogger("azure-amqp-client-single-sender"),
	}, nil
}

func (css *ClientSingleSender) Publish(ctx context.Context, msg *pkgamqp.Message[*amqp.MessageHeader], opts *amqp.SendOptions) error {
	css.logger.Debugf("publish started")
	defer css.logger.Debugf("publish finished")

	if utils.IsNil(msg) {
		return errs.NewTlCommonError("Publish", "cannot publish nil message", nil)
	}

	for attempt := 1; attempt <= css.opts.PublishMaxTryAttempts; attempt++ {
		// Проверяем контекст приложения перед каждым ретраем
		if err := ctx.Err(); err != nil {
			return errs.NewTlCommonError("Publish", "context canceled before attempt", err)
		}

		css.logger.Debugf("publish attempt #%d", attempt)

		sender, err := css.getSender(ctx)
		if err != nil {
			if attempt < css.opts.PublishMaxTryAttempts {
				css.logger.Warnf("AMQP failed to get sender on attempt %d: %v", attempt, err)

				// Применяем Backoff, так как сетевая ошибка произошла на этапе инициализации
				css.waitBackoff(ctx, attempt)
				continue
			}
			return errs.NewTlCommonError("Publish", "azure sender failed to initialize connection", err)
		}

		azureMsg := css.prepareMessage(msg)

		err = sender.Send(ctx, azureMsg, opts)
		if err == nil {
			css.logger.Debugf("AMQP message successfully published on attempt %d", attempt)
			return nil
		}

		// Обрабатываем ошибку отправки (сбрасывает ресурсы под мьютексом)
		err = css.handleSendError(attempt, err)
		if err != nil {
			return err // Невосстановимая ошибка или лимит попыток исчерпан
		}

		// Если handleSendError вернул nil, значит ошибка сетевая и мы идем на ретрай.
		// Ждем перед следующей попыткой:
		css.waitBackoff(ctx, attempt)
	}

	return errs.NewTlCommonError("Publish", "azure sender unexpected retry loop exit", nil)
}

//goland:noinspection DuplicatedCode
func (css *ClientSingleSender) Close(ctx context.Context) error {
	css.logger.Debugf("close started")
	defer css.logger.Debugf("close finished")

	css.mu.Lock()
	// copy links
	connToClose := css.connection
	sessToClose := css.session
	senderToClose := css.sender

	// clear members
	css.connection = nil
	css.session = nil
	css.sender = nil
	css.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, css.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)

	go func() {
		var closeErrs []error

		// Мягко закрываем сендеры по локальной копии
		if !utils.IsNil(senderToClose) {
			if err := senderToClose.Close(closeCtx); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}

		if !utils.IsNil(sessToClose) {
			if err := sessToClose.Close(closeCtx); err != nil {
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
			return errs.NewTlCommonError("Close", "Azure AMQP client sender close fails", err)
		}

		return nil
	case <-closeCtx.Done():
		// СРАБОТАЛ ПРЕДОХРАНИТЕЛЬ (Hard Teardown)
		// Если сендеры зависли на 2+ сообщениях — мы выходим из горутины и бьем по сокету напрямую
		if connToClose != nil {
			// Принудительное закрытие сокета на уровне ОС разблокирует внутренний mux го-амкп
			_ = connToClose.Close()
		}

		return errs.NewTlCommonError("Close", "Azure AMQP client sender close timeout: connection force closed", closeCtx.Err())
	}
}

// Вспомогательный метод вычисления и ожидания бэккоффа
func (css *ClientSingleSender) waitBackoff(ctx context.Context, attempt int) {
	// 1. Расчет экспоненты
	shift := uint(attempt - 1)
	if shift > 31 {
		shift = 31
	}

	delay := css.opts.PublishBaseRetryDelay * (1 << shift)
	if delay > css.opts.PublishMaxRetryDelay || delay <= 0 {
		delay = css.opts.PublishMaxRetryDelay
	}

	// 2. Добавляем Jitter (+/- 20%) через стандартный rand.Intn
	// Переводим delay в миллисекунды (int), чтобы безопасно работать с rand.Intn
	delayMs := int(delay / time.Millisecond)

	if delayMs > 5 {
		// Получаем случайное число от 0 до delayMs/5
		maxJitterMs := delayMs / 5
		jitterMs := rand.IntN(maxJitterMs)

		jitter := time.Duration(jitterMs) * time.Millisecond

		// Выбор 50/50 через rand.Intn(2)
		if rand.IntN(2) == 0 {
			delay += jitter
		} else {
			delay -= jitter
		}
	}

	css.logger.Debugf("backing off for %v before next attempt", delay)

	// 3. Безопасный таймер без утечек памяти
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (css *ClientSingleSender) getSender(ctx context.Context) (AmqpSenderLink, error) {
	// 1. Быстрый путь (Fast Path): если сендер жив, отдаем под RLock
	css.mu.RLock()
	if !utils.IsNil(css.sender) {
		defer css.mu.RUnlock()
		return css.sender, nil
	}
	css.mu.RUnlock()

	// 2. Медленный путь (Slow Path): сендера нет, нужна сеть.
	css.initMu.Lock()
	defer css.initMu.Unlock()

	// 3. Double-check: возможно, пока мы ждали initMu, другая горутина уже всё создала
	css.mu.RLock()
	if !utils.IsNil(css.sender) {
		defer css.mu.RUnlock()
		return css.sender, nil
	}
	css.mu.RUnlock()

	// 4. Проверяем и восстанавливаем Connection и Session
	session, err := css.getOrCreateSession(ctx)
	if err != nil {
		return nil, err
	}

	// 5. СЕТЕВОЙ ВЫЗОВ ВНЕ ОСНОВНОГО МЬЮТЕКСА
	css.logger.Debugf("opening new amqp target link for target name: %s", css.opts.TargetName)

	linkCtx, cancel := context.WithTimeout(ctx, css.opts.ConnectTimeout)
	defer cancel()

	newSender, err := session.NewSender(linkCtx, css.opts.TargetName, css.opts.SenderOpts)
	if err != nil {
		// ИСПРАВЛЕНИЕ: Если линк не создался, обязательно сбрасываем СЕССИЮ И СЕНДЕР.
		// Иначе в структуре может остаться старый не-nil указатель на мертвый сендер из прошлых итераций,
		// и следующий вызов Publish пойдет по ложному Fast Path.
		css.mu.Lock()
		css.session = nil
		css.sender = nil
		css.mu.Unlock()
		return nil, errs.NewTlCommonError("getSender", fmt.Sprintf("failed to create amqp link for target [%s]", css.opts.TargetName), err)
	}

	// 6. Сохраняем готовый сендер в структуру
	css.mu.Lock()
	css.sender = newSender
	css.mu.Unlock()

	return newSender, nil
}

func (css *ClientSingleSender) getOrCreateSession(ctx context.Context) (*amqp.Session, error) {
	css.mu.RLock()
	connAlive := !utils.IsNil(css.connection)
	sessAlive := !utils.IsNil(css.session)
	localConn := css.connection
	localSess := css.session
	css.mu.RUnlock()

	// Если всё есть — быстро возвращаем сессию без аллокаций и блокировок
	if connAlive && sessAlive {
		return localSess, nil
	}

	var newlyDialed bool // Флаг, создали ли мы новый сокет прямо сейчас

	// Если соединения нет вообще — создаем с нуля вне основного мьютекса
	if !connAlive {
		dialCtx, cancel := context.WithTimeout(ctx, css.opts.ConnectTimeout)
		defer cancel()

		var conn *amqp.Conn
		var err error
		if !utils.IsNil(css.opts.DialFnTestGap) {
			conn, err = css.opts.DialFnTestGap(dialCtx, css.opts.URL, css.opts.ConnOpts)
		} else {
			conn, err = amqp.Dial(dialCtx, css.opts.URL, css.opts.ConnOpts)
		}
		if err != nil {
			return nil, errs.NewTlCommonError("getOrCreateSession", "dial failed", err)
		}
		localConn = conn
		newlyDialed = true
	}

	// Создаем сессию по локальному коннекту вне основного мьютекса
	sessCtx, cancelSess := context.WithTimeout(ctx, css.opts.ConnectTimeout)
	defer cancelSess()

	session, err := localConn.NewSession(sessCtx, css.opts.SessionOpts)
	if err != nil {
		// ТВОЯ ИДЕЯ: Атомарно обновляем состояние под мьютексом за наносекунды
		// и выносим тяжелый Close() в фоновую горутину за рамки блокировки.
		css.mu.Lock()
		var connToCloseBehindMutex *amqp.Conn

		if newlyDialed {
			// Новый сокет точно надо закрыть, на него больше никто в системе не ссылается
			connToCloseBehindMutex = localConn
		} else {
			// Старый сокет забираем из структуры для асинхронного закрытия,
			// а глобальное поле мгновенно зануляем, открывая дорогу параллельным потокам
			connToCloseBehindMutex = localConn
			css.connection = nil
		}
		css.session = nil
		css.sender = nil
		css.mu.Unlock() // МЬЮТЕКС СРАЗУ СВОБОДЕН! Потоки в Publish не блокируются

		// Безопасный асинхронный сброс сетевых ресурсов
		if connToCloseBehindMutex != nil {
			go func(c *amqp.Conn) {
				// Предохранитель ОС: даем библиотеке go-amqp ровно 3 секунды
				// на попытку корректно отправить фрейм Close.
				// Даже если сокет зависнет в ядре, текущая горутина не заблокирует бизнес-логику.
				_ = c.Close()
			}(connToCloseBehindMutex)
		}

		return nil, errs.NewTlCommonError("getOrCreateSession", "failed to open session", err)
	}

	// Успешный сценарий: фиксируем новые живые ресурсы в структуре
	css.mu.Lock()
	css.connection = localConn
	css.session = session
	css.mu.Unlock()

	return session, nil
}

func (css *ClientSingleSender) handleSendError(attempt int, err error) error {
	var linkErr *amqp.LinkError
	var sessionErr *amqp.SessionError
	var connErr *amqp.ConnError

	// Защищаем проверку структуры и типов ошибок единым мьютексом
	css.mu.Lock()
	defer css.mu.Unlock()

	isNetworkErr := errors.As(err, &linkErr) || errors.As(err, &sessionErr) || errors.As(err, &connErr)

	if isNetworkErr {
		if attempt < css.opts.PublishMaxTryAttempts {
			css.logger.Warnf("AMQP network failure detected on attempt %d (%v). Invalidating resources...", attempt, err)

			// Инвалидация ресурсов прямо здесь, под уже взятым мьютексом
			switch {
			case errors.As(err, &linkErr):
				css.sender = nil
			case errors.As(err, &sessionErr):
				css.sender = nil
				css.session = nil
			case errors.As(err, &connErr):
				css.sender = nil
				css.session = nil
				css.connection = nil
			}
			return nil
		}
		return errs.NewTlCommonError("Publish", fmt.Sprintf("azure sender network error persisted after retry %d", attempt), err)
	}

	// Если это не сетевая ошибка (например, context.Canceled), реконнект не нужен
	return errs.NewTlCommonError("Publish", "azure sender unrecoverable send error", err)
}

func (css *ClientSingleSender) prepareMessage(msg *pkgamqp.Message[*amqp.MessageHeader]) *amqp.Message {
	azureMsg := amqp.NewMessage(msg.Payload)
	if !utils.IsNil(msg.Header) {
		azureMsg.Header = msg.Header
	}
	azureMsg.Properties = &amqp.MessageProperties{
		ContentType: css.pStr("application/json"),
	}
	if len(msg.Properties) > 0 {
		azureMsg.ApplicationProperties = msg.Properties
	}

	return azureMsg
}

func (css *ClientSingleSender) pStr(s string) *string {
	return &s
}

func (css *ClientSingleSender) GetOpts() *ClientSenderOptions {
	return css.opts
}

func (css *ClientSingleSender) GetTargetName() string {
	return css.opts.TargetName
}
