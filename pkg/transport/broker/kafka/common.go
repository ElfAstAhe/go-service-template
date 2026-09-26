package kafka

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
)

//goland:noinspection GoNameStartsWithPackageName
type KafkaSenderLink interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

//goland:noinspection GoNameStartsWithPackageName
type KafkaReceiverLink interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
	Stats() kafka.ReaderStats
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
