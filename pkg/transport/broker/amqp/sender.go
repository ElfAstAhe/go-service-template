package amqp

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
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

var jsonContentType = "application/json"

type Sender struct {
	opts   *SenderOptions
	sender AMQPSenderLink // Единственный фиксированный линк-отправитель
	logger logger.Logger
	mu     sync.RWMutex
	initMu sync.Mutex // Защищает ленивую инициализацию линка от Thundering Herd
}

// Привязываем структуру к итоговому интерфейсу пакета абстракций
var _ pkgamqp.Sender = (*Sender)(nil)
var _ AMQPSender = (*Sender)(nil)

func NewSender(opts ...SenderOption) (*Sender, error) {
	clientOpts := NewSenderOptions() // Все базовые дефолты таймаутов и бэккоффов внутри

	for _, opt := range opts {
		opt(clientOpts)
	}

	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewSender", "client sender options validate failed", err)
	}

	return &Sender{
		opts:   clientOpts,
		logger: clientOpts.Logger.GetLogger("azure-amqp-sender"),
	}, nil
}

func (s *Sender) Publish(ctx context.Context, msg pkgamqp.Message) error {
	return s.PublishWithOpts(ctx, msg, s.getSendOpts())
}

func (s *Sender) PublishWithOpts(ctx context.Context, msg pkgamqp.Message, sendOpts *amqp.SendOptions) error {
	s.logger.Debug("publish started")
	defer s.logger.Debug("publish finished")

	if utils.IsNil(msg) {
		return errs.NewTlCommonError("Publish", "cannot publish nil message", nil)
	}

	// опции
	opts := sendOpts
	if utils.IsNil(opts) {
		opts = s.getSendOpts()
	}

	for attempt := 1; attempt <= s.opts.PublishMaxTryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return errs.NewTlCommonError("Publish", "context canceled before attempt", err)
		}

		//if s.logger.IsDebugEnabled() {
		s.logger.Debugf("publish attempt #%d", attempt)
		//}

		// Получаем или лениво инициализируем линк очереди/топика
		senderLink, err := s.getSender(ctx)
		if err != nil {
			if attempt < s.opts.PublishMaxTryAttempts {
				s.logger.Warnf("AMQP failed to get sender on attempt %d: %v", attempt, err)
				s.waitBackoff(ctx, attempt)
				continue
			}

			return errs.NewTlCommonError("Publish", "azure sender failed to initialize link", err)
		}

		azureMsg := s.prepareMessage(msg)

		err = senderLink.Send(ctx, azureMsg, opts)
		if err == nil {
			s.logger.Debugf("AMQP message successfully published on attempt %d", attempt)

			return nil
		}

		// Обрабатываем сетевую ошибку — сообщаем общему коннектору для атомарного сброса
		err = s.handleSendError(attempt, err)
		if err != nil {
			return err
		}

		s.waitBackoff(ctx, attempt)
	}

	return errs.NewTlCommonError("Publish", "azure sender unexpected retry loop exit", nil)
}

//goland:noinspection DuplicatedCode
func (s *Sender) Close(ctx context.Context) error {
	s.logger.Debug("close started")
	defer s.logger.Debug("close finished")

	s.mu.Lock()
	senderToClose := s.sender
	s.sender = nil // Мгновенно обнуляем ссылку
	s.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)

	go func() {
		var closeErrs []error
		// Мягко гасим исключительно СВОЙ линк.
		// Сессию и коннект не трогаем — за них отвечает глобальный Connector.
		if !utils.IsNil(senderToClose) {
			if err := senderToClose.Close(closeCtx); err != nil {
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
		return errs.NewTlCommonError("Close", "Azure AMQP client sender close timeout", closeCtx.Err())
	}
}

func (s *Sender) GetTargetName() string {
	return s.opts.TargetName
}

//goland:noinspection DuplicatedCode
func (s *Sender) getSender(ctx context.Context) (AMQPSenderLink, error) {
	// 1. Быстрый путь (Fast Path): если линк жив, отдаем под RLock
	s.mu.RLock()
	if !utils.IsNil(s.sender) {
		s.mu.RUnlock()

		return s.sender, nil
	}
	s.mu.RUnlock()

	// 2. Медленный путь (Slow Path) под мьютексом инициализации линка
	s.initMu.Lock()
	defer s.initMu.Unlock()

	// 3. Double-check
	s.mu.RLock()
	if !utils.IsNil(s.sender) {
		s.mu.RUnlock()

		return s.sender, nil
	}
	s.mu.RUnlock()

	// 4. Запрашиваем живую сессию у общего коннектора через GetConnection
	session, err := s.opts.Connector.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Debugf("opening new amqp target link for target name: %s", s.opts.TargetName)

	linkCtx, cancel := context.WithTimeout(ctx, s.opts.ConnectTimeout)
	defer cancel()

	newSender, err := session.NewSender(linkCtx, s.opts.TargetName, s.opts.Opts)
	if err != nil {
		// Ошибка создания линка — сообщаем коннектору. Возможно, протухла сессия.
		s.opts.Connector.Invalidate(err)
		s.mu.Lock()
		s.sender = nil
		s.mu.Unlock()

		return nil, errs.NewTlCommonError("getSender", fmt.Sprintf("failed to create amqp link for target [%s]", s.opts.TargetName), err)
	}

	s.mu.Lock()
	s.sender = newSender
	s.mu.Unlock()

	return newSender, nil
}

// handleSendError анализирует причину сбоя при отправке сообщения.
// Если ошибка сетевая (Link/Session/Connection), метод атомарно сбрасывает локальный линк
// и делегирует инвалидацию общему Connector, подготавливая ленивый реконнект для следующей попытки.
func (s *Sender) handleSendError(attempt int, err error) error {
	var linkErr *amqp.LinkError
	var sessionErr *amqp.SessionError
	var connErr *amqp.ConnError

	// Кастим ошибку ко всем возможным уровням сетевых сбоев спецификации AMQP 1.0
	isNetworkErr := errors.As(err, &linkErr) || errors.As(err, &sessionErr) || errors.As(err, &connErr)

	if isNetworkErr {
		if attempt < s.opts.PublishMaxTryAttempts {
			s.logger.Warnf("AMQP network failure detected on attempt %d (%v). Notifying connector...", attempt, err)

			// Оповещаем глобальный коннектор. Он сам определит масштаб бедствия (сессия или весь сокет)
			s.opts.Connector.Invalidate(err)

			s.mu.Lock()
			s.sender = nil // Обнуляем локальный линк, провоцируя ленивое пересоздание в getSender
			s.mu.Unlock()
			return nil // Возвращаем nil, разрешая циклу Publish пойти на следующий ретрай
		}

		return errs.NewTlCommonError("Publish", fmt.Sprintf("azure sender network error persisted after retry %d", attempt), err)
	}

	// Любая не-сетевая ошибка (например, нарушение прав или битый payload) считается фатальной
	return errs.NewTlCommonError("Publish", "azure sender unrecoverable send error", err)
}

// waitBackoff рассчитывает паузу перед повторной попыткой по формуле экспоненциального бэкоффа
// и подмешивает к ней случайный джиттер (+/- 20%). Это размывает пиковую нагрузку на брокер
// от множества упавших подов (защита от Thundering Herd эффекта).
//
//goland:noinspection DuplicatedCode
func (s *Sender) waitBackoff(ctx context.Context, attempt int) {
	// Сдвиг битов ограничен, чтобы избежать переполнения типов (Overflow)
	shift := min(uint(attempt-1), 31)

	// Формула: BaseDelay * 2^(attempt-1)
	delay := s.opts.PublishBaseRetryDelay * (1 << shift)
	if delay > s.opts.PublishMaxRetryDelay || delay <= 0 {
		delay = s.opts.PublishMaxRetryDelay
	}

	delayMs := int(delay / time.Millisecond)

	// Если задержка существенна, подмешиваем 20%-й случайный джиттер
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

	s.logger.Debugf("backing off for %v before next attempt", delay)

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C: // Пауза успешно выдержана
	case <-ctx.Done(): // Контекст отменился до истечения таймера (плавный выход)
	}
}

// prepareMessage упаковывает абстрактное сообщение фреймворка в нативный конверт библиотеки go-amqp.
// Переносит payload, системные ApplicationProperties и сохраняет нативные AMQP заголовки (Header).
func (s *Sender) prepareMessage(msg pkgamqp.Message) *amqp.Message {
	azureMsg := amqp.NewMessage(msg.GetPayload())
	azureMsg.Properties = &amqp.MessageProperties{
		ContentType: &jsonContentType,
	}

	// Маппим пользовательские свойства (метаданные/заголовки)
	if len(msg.GetProperties()) > 0 {
		azureMsg.ApplicationProperties = msg.GetProperties()
	}

	// Если пришел наш родной конверт Azure — извлекаем и сохраняем низкоуровневые AMQP-заголовки
	if msgImpl, ok := msg.(*Message); ok {
		if !utils.IsNil(msgImpl.Header) {
			azureMsg.Header = msgImpl.Header
		}
	}

	return azureMsg
}

// getSendOpts возвращает кастомные опции публикации, либо откатывается на дефолты брокера.
func (s *Sender) getSendOpts() *amqp.SendOptions {
	if utils.IsNil(s.opts.SendOpts) {
		return s.buildDefaultOpts()
	}

	return s.opts.SendOpts
}

// buildDefaultOpts формирует безопасные дефолты публикации.
// Выставляет Settled = true (режим At-Most-Once / Fire-and-Forget) для максимальной пропускной способности.
func (s *Sender) buildDefaultOpts() *amqp.SendOptions {
	return &amqp.SendOptions{
		Settled: true,
	}
}
