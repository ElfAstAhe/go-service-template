package amqp

import (
	"context"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// SenderLink описывает методы встроенного отправителя библиотеки Azure AMQP,
// которые нам нужны для управления его жизненным циклом.
//
// Абстрагирует прикладной код от конкретной реализации продюсера библиотеки go-amqp.
//
//goland:noinspection GoNameStartsWithPackageName
type SenderLink interface {
	// Send осуществляет публикацию низкоуровневого сообщения msg в брокер с опциями opts.
	Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error

	// Close плавно закрывает активный линк отправки сообщений в рамках контекста ctx.
	Close(ctx context.Context) error
}

// ReceiverLink описывает методы встроенного получателя Azure AMQP,
// необходимые для чтения, подтверждения и закрытия линка.
//
// Абстрагирует прикладной код от конкретной реализации консьюмера библиотеки go-amqp.
//
//goland:noinspection GoNameStartsWithPackageName
type ReceiverLink interface {
	// Receive выполняет извлечение сообщения из AMQP-линка с заданными опциями.
	Receive(ctx context.Context, opts *amqp.ReceiveOptions) (*amqp.Message, error)

	// AcceptMessage подтверждает успешную обработку (positive acknowledgment/settle) сообщения брокером.
	AcceptMessage(ctx context.Context, msg *amqp.Message) error

	// RejectMessage отклоняет сообщение (negative acknowledgment) с фиксацией ошибки, отправляя его в DLQ.
	RejectMessage(ctx context.Context, msg *amqp.Message, err *amqp.Error) error

	// ReleaseMessage освобождает сообщение, возвращая его обратно в очередь для повторной вычитки другими потоками.
	ReleaseMessage(ctx context.Context, msg *amqp.Message) error

	// Close плавно закрывает активный линк получения сообщений в рамках контекста ctx.
	Close(ctx context.Context) error
}

// ExtractOriginalMessage выполняет приведение полиморфного интерфейса сообщения фреймворка
// к низкоуровневой структуре amqp.Message драйвера Azure go-amqp.
// Защищает рантайм от паник времени выполнения посредством многоуровневых оборонительных проверок (Guard Clauses).
func ExtractOriginalMessage(msg broker.Message) (*amqp.Message, error) {
	if utils.IsNil(msg) {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "message is nil", nil)
	}
	raw, err := msg.ExtractOriginalMessage()
	if err != nil {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "extraction failed", err)
	}

	return raw.(*amqp.Message), nil
}
