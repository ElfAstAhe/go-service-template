package azure

import (
	"context"

	"github.com/Azure/go-amqp"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
)

// AMQPReceiver — расширенный локальный интерфейс для Azure Service Bus.
// Успешно встраивает плоский контракт фреймворка и добавляет метод ручного управления опциями.
type AMQPReceiver interface {
	pkgamqp.Receiver

	ReceiveWithOpts(ctx context.Context, opts *amqp.ReceiveOptions) (pkgamqp.Message, error)
}
