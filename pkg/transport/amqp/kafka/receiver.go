package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"sync"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// Receiver реализует интерфейс транспортного уровня для вычитки сообщений из Kafka.
// Поддерживает работу в рамках распределенной Consumer Group и ручное управление смещениями.
type Receiver struct {
	opts        *ReceiverOptions    // Опции рантайма, переданные при создании
	reader      KafkaReceiverLink   // Интерфейс обертки над низкоуровневым *kafka.Reader
	logger      logger.Logger       // Системный логгер приложения
	kafkaLogger *logger.KafkaLogger // Адаптер для перехвата внутренних логов библиотеки kafka-go
	mu          sync.RWMutex        // Мьютекс для потокобезопасного чтения/записи активного ридера
	initMu      sync.Mutex          // Мьютекс для защиты от Race Condition при инициализации (Double-Checked Locking)
}

// Гарантируем соответствие общему интерфейсу amqp.Receiver на этапе компиляции.
var _ pkgamqp.Receiver[any] = (*Receiver)(nil)

// NewReceiver — конструктор компонента Receiver. Накатывает переданные функциональные опции,
// инициализирует логгеры и возвращает готовый к работе экземпляр.
func NewReceiver(opts ...ReceiverOption) (*Receiver, error) {
	clientOpts := NewReceiverOptions()
	for _, opt := range opts {
		opt(clientOpts)
	}
	// Проверяем валидность опций, включая золотое правило соотношения сессии и хартбитов
	if err := clientOpts.Validate(); err != nil {
		return nil, errs.NewTlCommonError("NewReceiver", "kafka receiver options validation failed", err)
	}

	log := clientOpts.Logger.GetLogger("kafka-receiver")

	return &Receiver{
		opts:        clientOpts,
		logger:      log,
		kafkaLogger: logger.NewKafkaLogger(log),
	}, nil
}

// Receive блокирует текущий поток и дожидается поступления нового сообщения из Kafka.
// Возвращает независимый конверт сообщения pkgamqp.Message.
func (r *Receiver) Receive(ctx context.Context, _ any) (pkgamqp.Message, error) {
	// Получаем или лениво инициализируем живой инстанс ридера
	receiverLink, err := r.getReceiver(ctx)
	if err != nil {
		return nil, errs.NewTlCommonError("Receive", "kafka receiver failed to get reader", err)
	}

	// FetchMessage блокирует вызов до прихода пачки данных из сети и НЕ коммитит оффсет автоматически
	kafkaMsg, err := receiverLink.FetchMessage(ctx)
	if err != nil {
		return nil, errs.NewTlCommonError("Receive", "kafka incoming packet error", err)
	}

	// Безопасно изолируем payload сообщения, выделяя под него чистую память
	var finalPayload []byte
	if len(kafkaMsg.Value) > 0 {
		finalPayload = make([]byte, len(kafkaMsg.Value))
		copy(finalPayload, kafkaMsg.Value)
	}

	// Собираем наш универсальный конверт с помощью протестированной фабрики NewMessage
	resMsg := NewMessage(kafkaMsg)
	resMsg.TargetName = r.opts.TargetName

	// Дополнительно прокидываем строковое представление ключа партиционирования в Props
	if len(kafkaMsg.Key) > 0 {
		resMsg.Props["kafka_message_key"] = string(kafkaMsg.Key)
	}

	return resMsg, nil
}

// Accept подтверждает брокеру успешную обработку кадра сообщения, фиксируя его смещение (Commit Offset) в Kafka.
func (r *Receiver) Accept(ctx context.Context, msg pkgamqp.Message) error {
	// Извлекаем указатель на оригинальный пакет *kafka.Message с помощью нашего безопасного хелпера
	kafkaMsgPtr, err := ExtractOriginalKafkaMessage(msg)
	if err != nil {
		return err
	}

	kafkaReader, err := r.getReceiver(ctx)
	if err != nil {
		return err
	}

	// Синхронно фиксируем прогресс чтения для данной партиции в кластере Kafka
	if err = kafkaReader.CommitMessages(ctx, *kafkaMsgPtr); err != nil {
		return errs.NewTlCommonError("Accept", "kafka failed to commit offset", err)
	}
	return nil
}

// Reject обрабатывает "битые" кадры или сообщения, завершившиеся критической бизнес-ошибкой.
// В Kafka мы продвигаем смещение вперед (вызывая Accept), чтобы не заблокировать поток (Head-of-line blocking).
func (r *Receiver) Reject(ctx context.Context, msg pkgamqp.Message, err error) error {
	r.logger.Errorf("Message processing rejected: %v. Moving offset forward.", err)
	return r.Accept(ctx, msg)
}

// Release обрабатывает временные сбои (например, моргание БД). В нашей конфигурации ручного управления оффсетами
// мы просто ничего не делаем. Сообщение будет прочитано заново после ребалансировки или перезапуска пода.
func (r *Receiver) Release(ctx context.Context, msg pkgamqp.Message) error {
	r.logger.Warnf("Kafka release called: message will be re-read upon partition rebalance.")
	return nil
}

// Close мягко останавливает консьюмер, завершая сетевые сессии и выходя из Consumer Group брокера.
// Реализована защита от зависания по таймауту ShutdownTimeout.
//
//goland:noinspection DuplicatedCode
func (r *Receiver) Close(ctx context.Context) error {
	r.mu.Lock()
	readerToClose := r.reader
	r.reader = nil // Обнуляем ссылку, чтобы предотвратить параллельные вызовы Fetch во время закрытия
	r.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, r.opts.ShutdownTimeout)
	defer closeCancel()

	done := make(chan error, 1)
	go func() {
		var closeErrs []error
		if !utils.IsNil(readerToClose) {
			if err := readerToClose.Close(); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}
		done <- errors.Join(closeErrs...)
	}()

	select {
	case err := <-done:
		if err != nil {
			return errs.NewTlCommonError("Close", "kafka receiver close fails", err)
		}
		return nil
	case <-closeCtx.Done():
		return errs.NewTlCommonError("Close", "kafka receiver close timeout", closeCtx.Err())
	}
}

// GetTargetName возвращает имя топика Kafka, который слушает данный Receiver.
func (r *Receiver) GetTargetName() string { return r.opts.TargetName }

// getReceiver инициализирует или возвращает существующий линк ридера (Double-Checked Locking паттерн).
func (r *Receiver) getReceiver(ctx context.Context) (KafkaReceiverLink, error) {
	// Первая быстрая проверка под RLock (Fast Path)
	r.mu.RLock()
	if !utils.IsNil(r.reader) {
		defer r.mu.RUnlock()
		return r.reader, nil
	}
	r.mu.RUnlock()

	// Эксклюзивная блокировка на создание объекта (Slow Path)
	r.initMu.Lock()
	defer r.initMu.Unlock()

	// Вторая проверка под RLock на случай, если параллельный поток успел создать ридер, пока мы ждали initMu
	r.mu.RLock()
	if !utils.IsNil(r.reader) {
		defer r.mu.RUnlock()
		return r.reader, nil
	}
	r.mu.RUnlock()

	// Создаем сам объект пуллера
	newReader := kafka.NewReader(r.createReaderConfig(r.createDealer()))

	r.mu.Lock()
	r.reader = newReader
	r.mu.Unlock()

	return newReader, nil
}

func (r *Receiver) getDealerTLS() *tls.Config {
	if !utils.IsNil(r.opts.TLS) {
		return r.opts.TLS
	}

	return nil
}

func (r *Receiver) createDealer() *kafka.Dialer {
	// Настраиваем сетевой Dialer сокета (уровень TCP/TLS соединений)
	dialer := &kafka.Dialer{
		Timeout:   r.opts.ConnectTimeout,
		DualStack: true,
		TLS:       r.getDealerTLS(),
	}

	// Если в конфигурации передан Username, инициализируем безопасный SASL слой поверх TLS
	if strings.TrimSpace(r.opts.Username) != "" {
		dialer.SASLMechanism = plain.Mechanism{
			Username: r.opts.Username,
			Password: r.opts.Password,
		}
	}

	return dialer
}

func (r *Receiver) createReaderConfig(dialer *kafka.Dialer) kafka.ReaderConfig {
	// Маппим строковую политику точки старта в системную константу типа int64 библиотеки kafka-go
	var startOffset = kafka.FirstOffset
	if strings.ToLower(r.opts.StartOffset) == "last" {
		startOffset = kafka.LastOffset
	}

	// Формируем базовую конфигурацию ReaderConfig, утилизируя все наши новые таймауты координации
	readerCfg := kafka.ReaderConfig{
		Brokers:           r.opts.Brokers,
		Topic:             r.opts.TargetName,
		GroupID:           r.opts.GroupID,
		Dialer:            dialer,
		MinBytes:          r.opts.MinBytes,
		MaxBytes:          r.opts.MaxBytes,
		MaxWait:           r.opts.MaxWait,
		HeartbeatInterval: r.opts.HeartbeatInterval, // Фоновый пинг брокера
		SessionTimeout:    r.opts.SessionTimeout,    // Время жизни сессии воркера
		RebalanceTimeout:  r.opts.RebalanceTimeout,  // Время на сдачу оффсетов при ребалансировке групп
		ReadBatchTimeout:  r.opts.ReadBatchTimeout,  // Сетевой таймаут сокета на вычитку одного батча
		MaxAttempts:       r.opts.MaxAttempts,       // Количество попыток сетевых ретраев
		QueueCapacity:     r.opts.QueueCapacity,     // Размер внутреннего фонового буфера сообщений
		StartOffset:       startOffset,              // Точка старта при отсутствии оффсетов
		CommitInterval:    0,                        // Переводим коммиты строго в синхронный ручной режим
		Logger:            r.kafkaLogger.InfoLogger(),
		ErrorLogger:       r.kafkaLogger.ErrorLogger(),
	}

	// Если разработчик передал кастомный низкоуровневый ReaderConf, берем его за основу,
	// перетирая только критически важные для стабильности нашей обертки параметры
	if r.opts.ReaderConf != nil {
		readerCfg = *r.opts.ReaderConf
		readerCfg.Brokers = r.opts.Brokers
		readerCfg.Topic = r.opts.TargetName
		readerCfg.GroupID = r.opts.GroupID
		readerCfg.Dialer = dialer
		readerCfg.CommitInterval = 0
		readerCfg.Logger = r.kafkaLogger.InfoLogger()
		readerCfg.ErrorLogger = r.kafkaLogger.ErrorLogger()
	}

	return readerCfg
}
