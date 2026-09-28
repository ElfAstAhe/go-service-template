package amqp

import (
	"context"

	"github.com/Azure/go-amqp"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
)

// Receiver — расширенный локальный интерфейс для Azure Service Bus.
// Успешно встраивает плоский контракт фреймворка и добавляет метод ручного управления опциями.
//
// Выступает специализированным контрактом получателя (Consumer) для работы со спецификацией AMQP 1.0.
// Предоставляет прикладному слою возможность точечного тюнинга Flow Control и Prefetch политик.
//
//goland:noinspection GoNameStartsWithPackageName
type Receiver interface {
	pkgamqp.Receiver // Встраивает базовые методы вычитки и жизненного цикла консьюмера фреймворка

	// ReceiveWithOpts выполняет блокирующее извлечение сообщения из очереди брокера
	// с возможностью ручной передачи и настройки параметров сессии amqp.ReceiveOptions.
	ReceiveWithOpts(ctx context.Context, opts *amqp.ReceiveOptions) (pkgamqp.Message, error)
}
