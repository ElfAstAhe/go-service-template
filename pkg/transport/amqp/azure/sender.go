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

var jsonContentType = "application/json"

type Sender struct {
	opts   *SenderOptions
	sender AmqpSenderLink // Единственный фиксированный линк-отправитель
	logger logger.Logger
	mu     sync.RWMutex
	initMu sync.Mutex // Защищает ленивую инициализацию линка от Thundering Herd
}

// Привязываем структуру к итоговому интерфейсу пакета абстракций
var _ pkgamqp.Sender[*amqp.SendOptions, *amqp.MessageHeader] = (*Sender)(nil)

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

func (s *Sender) Publish(ctx context.Context, msg *pkgamqp.Message[*amqp.MessageHeader], opts *amqp.SendOptions) error {
	s.logger.Debug("publish started")
	defer s.logger.Debug("publish finished")

	if utils.IsNil(msg) {
		return errs.NewTlCommonError("Publish", "cannot publish nil message", nil)
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

func (s *Sender) getSender(ctx context.Context) (AmqpSenderLink, error) {
	// 1. Быстрый путь (Fast Path): если линк жив, отдаем под RLock
	s.mu.RLock()
	if !utils.IsNil(s.sender) {
		defer s.mu.RUnlock()
		return s.sender, nil
	}
	s.mu.RUnlock()

	// 2. Медленный путь (Slow Path) под мьютексом инициализации линка
	s.initMu.Lock()
	defer s.initMu.Unlock()

	// 3. Double-check
	s.mu.RLock()
	if !utils.IsNil(s.sender) {
		defer s.mu.RUnlock()

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

func (s *Sender) handleSendError(attempt int, err error) error {
	var linkErr *amqp.LinkError
	var sessionErr *amqp.SessionError
	var connErr *amqp.ConnError

	isNetworkErr := errors.As(err, &linkErr) || errors.As(err, &sessionErr) || errors.As(err, &connErr)

	if isNetworkErr {
		if attempt < s.opts.PublishMaxTryAttempts {
			s.logger.Warnf("AMQP network failure detected on attempt %d (%v). Notifying connector...", attempt, err)

			// ИСПРАВЛЕНИЕ: Передаем ошибку в общий коннектор. Он сам разберется, какой уровень инвалидировать.
			s.opts.Connector.Invalidate(err)

			s.mu.Lock()
			s.sender = nil // В любом случае сбрасываем локальный линк
			s.mu.Unlock()
			return nil
		}

		return errs.NewTlCommonError("Publish", fmt.Sprintf("azure sender network error persisted after retry %d", attempt), err)
	}

	return errs.NewTlCommonError("Publish", "azure sender unrecoverable send error", err)
}

func (s *Sender) waitBackoff(ctx context.Context, attempt int) {
	shift := uint(attempt - 1)
	if shift > 31 {
		shift = 31
	}

	delay := s.opts.PublishBaseRetryDelay * (1 << shift)
	if delay > s.opts.PublishMaxRetryDelay || delay <= 0 {
		delay = s.opts.PublishMaxRetryDelay
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

	s.logger.Debugf("backing off for %v before next attempt", delay)

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (s *Sender) prepareMessage(msg *pkgamqp.Message[*amqp.MessageHeader]) *amqp.Message {
	azureMsg := amqp.NewMessage(msg.Payload)
	if !utils.IsNil(msg.Header) {
		azureMsg.Header = msg.Header
	}
	azureMsg.Properties = &amqp.MessageProperties{
		ContentType: &jsonContentType,
	}
	if len(msg.Properties) > 0 {
		azureMsg.ApplicationProperties = msg.Properties
	}

	return azureMsg
}
