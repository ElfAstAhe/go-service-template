package kafka

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
)

const (
	DefaultConnectTimeout  time.Duration = 5 * time.Second
	DefaultShutdownTimeout time.Duration = 5 * time.Second
)

type KafkaSenderLink interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type KafkaReceiverLink interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

func ExtractOriginalKafkaMessage(msg pkgamqp.Message) (*kafka.Message, error) {
	if utils.IsNil(msg) {
		return nil, errs.NewTlCommonError("ExtractOriginalKafkaMessage", "message is nil", nil)
	}
	raw, err := msg.ExtractOriginalMessage()
	if err != nil {
		return nil, errs.NewTlCommonError("ExtractOriginalKafkaMessage", "extraction failed", err)
	}

	// Безопасное приведение к указателю
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
