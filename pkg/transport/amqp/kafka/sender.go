package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// Sender реализует отправку сообщений в конкретный топик Kafka.
// Поддерживает политики экспоненциального бэкоффа с джиттером.
type Sender struct {
	opts        *SenderOptions
	writer      KafkaSenderLink
	logger      logger.Logger
	kafkaLogger *logger.KafkaLogger
	mu          sync.RWMutex
	initMu      sync.Mutex
}

// Привязываем структуру к общему интерфейсу amqp.Sender.
var _ pkgamqp.Sender[any] = (*Sender)(nil)

// NewSender создает новый экземпляр отправителя на основе переданных опций.
//
//goland:noinspection GoUnusedExportedFunction
func NewSender(opts ...SenderOption) (*Sender, error) {
	clientOpts := NewSenderOptions()
	for _, opt := range opts {
		opt(clientOpts)
	}
	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewSender", "kafka sender options validation failed", err)
	}

	log := clientOpts.Logger.GetLogger("kafka-sender")

	return &Sender{
		opts:        clientOpts,
		logger:      log,
		kafkaLogger: logger.NewKafkaLogger(log),
	}, nil
}

// Publish отправляет сообщение в Kafka. Включает механизм повторных попыток при сетевых сбоях.
func (s *Sender) Publish(ctx context.Context, msg pkgamqp.Message, _ any) error {
	if utils.IsNil(msg) {
		return errs.NewTlCommonError("Publish", "cannot publish nil message", nil)
	}

	for attempt := 1; attempt <= s.opts.PublishMaxTryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return errs.NewTlCommonError("Publish", "context canceled", err)
		}

		kafkaWriter, err := s.getSender(ctx)
		if err != nil {
			if attempt < s.opts.PublishMaxTryAttempts {
				s.waitBackoff(ctx, attempt)
				continue
			}
			return errs.NewTlCommonError("Publish", "kafka sender failed to init writer", err)
		}

		kafkaMsg := s.prepareMessage(msg)

		err = s.internalWriteMessages(ctx, kafkaWriter, kafkaMsg)
		if err == nil {
			return nil // Успешная отправка
		}

		// Логируем ошибку и проверяем, имеет ли смысл делать ретрай
		if !s.isRecoverableError(err) || attempt == s.opts.PublishMaxTryAttempts {
			return errs.NewTlCommonError("Publish", "kafka unrecoverable send error or retries exhausted", err)
		}

		s.logger.Warnf("Temporary error sending to Kafka (attempt %d/%d): %v. Retrying...", attempt, s.opts.PublishMaxTryAttempts, err)
		s.waitBackoff(ctx, attempt)
	}

	return errs.NewTlCommonError("Publish", "kafka sender unexpected retry loop exit", nil)
}

// Close плавно завершает работу врайтера, дожидаясь отправки пакетов из буферов.
//
//goland:noinspection DuplicatedCode
func (s *Sender) Close(ctx context.Context) error {
	s.mu.Lock()
	writerToClose := s.writer
	s.writer = nil
	s.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, s.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)
	go func() {
		var closeErrs []error
		if !utils.IsNil(writerToClose) {
			if err := writerToClose.Close(); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}
		done <- errors.Join(closeErrs...)
	}()

	select {
	case err := <-done:
		if err != nil {
			return errs.NewTlCommonError("Close", "kafka sender close fails", err)
		}
		return nil
	case <-closeCtx.Done():
		return errs.NewTlCommonError("Close", "kafka sender close timeout", closeCtx.Err())
	}
}

// GetTargetName возвращает имя топика назначения.
func (s *Sender) GetTargetName() string {
	return s.opts.TargetName
}

// getSender инициализирует или возвращает существующий линк врайтера (Double-Checked Locking паттерн).
//
//goland:noinspection GoUnusedParameter,DuplicatedCode
func (s *Sender) getSender(ctx context.Context) (KafkaSenderLink, error) {
	// Первая быстрая проверка под RLock (Fast Path)
	s.mu.RLock()
	if !utils.IsNil(s.writer) {
		s.mu.RUnlock()
		return s.writer, nil
	}
	s.mu.RUnlock()

	// Эксклюзивная блокировка на создание объекта (Slow Path)
	s.initMu.Lock()
	defer s.initMu.Unlock()

	// Вторая проверка под RLock на случай, если параллельный поток успел создать врайтер, пока мы ждали initMu
	s.mu.RLock()
	if !utils.IsNil(s.writer) {
		s.mu.RUnlock() // ИСПРАВЛЕНО: Явный анлок вместо дефера защищает от вечной блокировки
		return s.writer, nil
	}
	s.mu.RUnlock()

	// Создаем сам объект врайтера на основе детерминированных фабричных методов
	newWriter := s.createWriter()

	s.mu.Lock()
	s.writer = newWriter
	s.mu.Unlock()

	return newWriter, nil
}

// createWriter собирает структуру kafka.Writer, утилизируя параметры асинхронного батчинга и кастомные мутаторы.
func (s *Sender) createWriter() *kafka.Writer {
	// Собираем инстанс напрямую через поля структуры (Modern-Way)
	newWriter := &kafka.Writer{
		Addr:         kafka.TCP(s.opts.Brokers...),
		Topic:        s.opts.TargetName,
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  1,                     // Повторами управляем сами в методе Publish с кастомным бэкоффом
		BatchSize:    s.opts.BatchSize,      // Лимит количества сообщений в пачке
		BatchBytes:   s.opts.BatchBytes,     // Лимит веса пачки в байтах (теперь честный int64)
		BatchTimeout: s.opts.BatchTimeout,   // Время ожидания сброса неполного батча
		WriteTimeout: s.opts.WriteTimeout,   // Сетевой таймаут сокета на запись пачки
		ReadTimeout:  s.opts.ConnectTimeout, // Таймаут на чтение ответа брокера
		RequiredAcks: s.opts.RequiredAcks,   // Уровень квитирования репликами (-1, 0, 1)
		Transport:    s.createTransport(),   // Слой сетевой безопасности (SASL/TLS)
		Logger:       s.kafkaLogger.InfoLogger(),
		ErrorLogger:  s.kafkaLogger.ErrorLogger(),
	}

	// Если конечная система передала функцию конфигурации,
	// скармливаем ей собранный объект. Она сможет переопределить любое поле (например, запустить Async: true).
	if !utils.IsNil(s.opts.WriterCustomizerFunc) {
		s.opts.WriterCustomizerFunc(newWriter)

		// Страхуем критически важные для стабильности нашей обертки параметры от случайной перезаписи
		newWriter.Addr = kafka.TCP(s.opts.Brokers...)
		newWriter.Topic = s.opts.TargetName
		newWriter.MaxAttempts = 1
		newWriter.Logger = s.kafkaLogger.InfoLogger()
		newWriter.ErrorLogger = s.kafkaLogger.ErrorLogger()
	}

	return newWriter
}

// createTransport настраивает сетевой транспортный слой, подставляя переданный TLS и накладывая SASL.
func (s *Sender) createTransport() *kafka.Transport {
	var transport = &kafka.Transport{
		DialTimeout: s.opts.ConnectTimeout,
		IdleTimeout: s.opts.IdleTimeout,
		ClientID:    s.opts.ClientID,
		TLS:         s.getTLS(),
	}

	// Если в конфигурации передан Username, накладываем поверх безопасный SASL слой
	if strings.TrimSpace(s.opts.Username) != "" {
		transport.SASL = plain.Mechanism{
			Username: s.opts.Username,
			Password: s.opts.Password,
		}
	}

	return transport // Вернет nil для Plaintext (локальной разработки на localhost)
}

// isRecoverableError определяет, является ли ошибка временной (сетевой), допуская повторную попытку.
func (s *Sender) isRecoverableError(err error) bool {
	var netErr net.Error
	var kErr kafka.Error

	return errors.As(err, &netErr) || errors.As(err, &kErr)
}

// waitBackoff вычисляет экспоненциальную задержку с добавлением случайного джиттера.
//
//goland:noinspection DuplicatedCode
func (s *Sender) waitBackoff(ctx context.Context, attempt int) {
	shift := min(uint(attempt-1), 31)
	delay := s.opts.PublishBaseRetryDelay * (1 << shift)
	if delay > s.opts.PublishMaxRetryDelay || delay <= 0 {
		delay = s.opts.PublishMaxRetryDelay
	}
	delayMs := int(delay / time.Millisecond)

	if delayMs > 5 {
		maxJitterMs := delayMs / 5
		jitterMs := rand.IntN(maxJitterMs) // Использование потокобезопасного v2 крипто-рандомайзера
		jitter := time.Duration(jitterMs) * time.Millisecond
		if rand.IntN(2) == 0 {
			delay += jitter
		} else {
			delay -= jitter
		}
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

// prepareMessage перекладывает полезную нагрузку и свойства нашего конверта в нативную структуру kafka.Message.
func (s *Sender) prepareMessage(msg pkgamqp.Message) kafka.Message {
	kafkaMsg := kafka.Message{Value: msg.GetPayload()}
	props := msg.GetProperties()
	if len(props) > 0 {
		// Извлекаем ключ партиционирования Kafka
		if keyVal, ok := props["kafka_message_key"]; ok {
			if strKey, ok := keyVal.(string); ok {
				kafkaMsg.Key = []byte(strKey)
			}
		}

		// Перекладываем все остальные свойства в Headers
		var headers []kafka.Header
		for k, v := range props {
			if k == "kafka_message_key" || k == sysKafkaMsgKey {
				continue // Игнорируем служебные ключи нашего пакета
			}
			if strVal, ok := v.(string); ok {
				headers = append(headers, kafka.Header{Key: k, Value: []byte(strVal)})
			}
		}
		kafkaMsg.Headers = headers
	}
	return kafkaMsg
}

func (s *Sender) getTLS() *tls.Config {
	if !utils.IsNil(s.opts.TLS) {
		return s.opts.TLS
	}

	return nil
}

// internalWriteMessages оборачивает вызов библиотеки в блок recover для перехвата рантайм-паник.
func (s *Sender) internalWriteMessages(ctx context.Context, writer KafkaSenderLink, messages ...kafka.Message) (err error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("Kafka writer critically panicked during WriteMessages: %v", r)
			// Превращаем панику в читаемую ошибку для логов
			var recoveryErr error
			if e, ok := r.(error); ok {
				recoveryErr = errs.NewCommonError("panic recovery", e)
			} else {
				recoveryErr = errs.NewCommonError(fmt.Sprintf("panic recovery [%v]", r), nil)
			}

			// Перехватываем панику, логируем её и превращаем в обычную ошибку компиляции/рантайма
			err = errs.NewTlCommonError("internalWriteMessages", fmt.Sprintf("kafka writer panic recovery: %v", r), recoveryErr)
		}
	}()

	// Вызываем нативный метод библиотеки, присваиваем результат вызова именованной переменной err перед возвратом
	err = writer.WriteMessages(ctx, messages...)

	return err
}
