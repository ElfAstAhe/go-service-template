package amqp

import (
	"context"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// AMQPSenderLink описывает методы встроенного отправителя библиотеки Azure AMQP,
// которые нам нужны для управления его жизненным циклом.
//
//goland:noinspection GoNameStartsWithPackageName
type AMQPSenderLink interface {
	Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error
	Close(ctx context.Context) error
}

// AMQPReceiverLink описывает методы встроенного получателя Azure AMQP,
// необходимые для чтения, подтверждения и закрытия линка.
//
//goland:noinspection GoNameStartsWithPackageName
type AMQPReceiverLink interface {
	Receive(ctx context.Context, opts *amqp.ReceiveOptions) (*amqp.Message, error)
	AcceptMessage(ctx context.Context, msg *amqp.Message) error
	RejectMessage(ctx context.Context, msg *amqp.Message, err *amqp.Error) error
	ReleaseMessage(ctx context.Context, msg *amqp.Message) error
	Close(ctx context.Context) error
}

func ExtractOriginalMessage(msg pkgamqp.Message) (*amqp.Message, error) {
	if utils.IsNil(msg) {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "message is nil", nil)
	}
	raw, err := msg.ExtractOriginalMessage()
	if err != nil {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "extraction failed", err)
	}

	return raw.(*amqp.Message), nil
}
