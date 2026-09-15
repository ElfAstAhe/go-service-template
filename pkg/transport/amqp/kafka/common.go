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
	sysMsgKey                            = "sys_original_kafka_message"
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

// Message реализует ваш интерфейс pkgamqp.Message
type Message struct {
	TargetName string
	Payload    []byte
	Props      map[string]any
}

var _ pkgamqp.Message = (*Message)(nil)

func (m *Message) GetTargetName() string         { return m.TargetName }
func (m *Message) GetPayload() []byte            { return m.Payload }
func (m *Message) GetProperties() map[string]any { return m.Props }
func (m *Message) ExtractOriginalMessage() (any, error) {
	if utils.IsNil(m.Props) {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "properties map is nil", nil)
	}
	raw, ok := m.Props[sysMsgKey]
	if !ok {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "metadata missing", nil)
	}
	return raw, nil
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
