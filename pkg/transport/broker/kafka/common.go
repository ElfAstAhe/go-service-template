package kafka

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
)

// KafkaSenderLink описывает изолированный, абстрактный контракт для низкоуровневого продюсера Apache Kafka.
//
// Инкапсулирует операции пакетной публикации сообщений и освобождения сетевых дескрипторов врайтера.
// Позволяет абстрагировать прикладной код от конкретной реализации (например, от структуры *kafka.Writer библиотеки segmentio).
//
//goland:noinspection GoNameStartsWithPackageName
type KafkaSenderLink interface {
	// WriteMessages выполняет атомарную пакетную запись сообщений в топики брокера Kafka.
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error

	// Close осуществляет плавное закрытие продюсера с предварительным принудительным сбросом (flushing) буферов из памяти на диски брокеров.
	Close() error
}

// KafkaReceiverLink описывает изолированный, абстрактный контракт для низкоуровневого консьюмера Apache Kafka.
//
// Инкапсулирует рутину поштучной вычитки сообщений (Fetch), коммита оффсетов и сбора рантайм-статистики читателя.
//
//goland:noinspection GoNameStartsWithPackageName
type KafkaReceiverLink interface {
	// FetchMessage извлекает очередное сообщение из партиции топика без автоматической фиксации оффсета.
	FetchMessage(ctx context.Context) (kafka.Message, error)

	// CommitMessages выполняет явное подтверждение (коммит) оффсетов для группы обработанных сообщений.
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error

	// Close осуществляет мягкий контролируемый выход консьюмера из Consumer Group брокера.
	Close() error

	// Stats экспортирует актуальный слепок внутренних метрик ридера (метрики лага, сетевых ошибок, RPS).
	Stats() kafka.ReaderStats
}

// ExtractOriginalKafkaMessage выполняет безопасное приведение полиморфного интерфейса сообщения фреймворка
// к низкоуровневой структуре kafka.Message драйвера segmentio.
// Защищает рантайм от паник времени выполнения посредством многоуровневых оборонительных проверок (Guard Clauses).
func ExtractOriginalKafkaMessage(msg pkgamqp.Message) (*kafka.Message, error) {
	// 1. Барьер валидации ссылки: пресекаем обработку пустых интерфейсных объектов
	if utils.IsNil(msg) {
		return nil, errs.NewTlCommonError("ExtractOriginalKafkaMessage", "message is nil", nil)
	}

	// 2. Извлекаем сырой нетипизированный non-exported объект-контейнер
	raw, err := msg.ExtractOriginalMessage()
	if err != nil {
		return nil, errs.NewTlCommonError("ExtractOriginalKafkaMessage", "extraction failed", err)
	}

	// 3. Безопасное приведение типов (Type Assertion) к целевому указателю драйвера segmentio
	kafkaMsgPtr, ok := raw.(*kafka.Message)
	if !ok {
		return nil, errs.NewTlCommonError(
			"ExtractOriginalKafkaMessage",
			"extracted message has invalid type, expected *kafka.Message",
			nil,
		)
	}

	return kafkaMsgPtr, nil
}
