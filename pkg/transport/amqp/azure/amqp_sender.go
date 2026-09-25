package azure

import (
	"context"

	"github.com/Azure/go-amqp"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
)

// AMQPSender — расширенный локальный интерфейс для отправки сообщений через AMQP 1.0.
//
// Использует паттерн Interface Extension (расширение интерфейса):
//  1. Встраивает базовый плоский контракт фреймворка pkgamqp.Sender для ультимативного полиморфизма
//     (бизнес-логика/обсерверы работают только с ним, не зная про кишки конкретного брокера).
//  2. Предоставляет специализированный метод PublishWithOpts для сценариев, где в рантайме
//     необходимо точечно переопределить нативные опции отправки библиотеки go-amqp.
type AMQPSender interface {
	pkgamqp.Sender // Встраиваем базовый плоский интерфейс фреймворка (GetTargetName, Publish)

	// PublishWithOpts выполняет отправку сообщения с явным указанием нативных опций Azure.
	// Используется для расширенных сценариев (например, динамическое выставление Message Delay,
	// Scheduled Delivery или ручная настройка системных заголовков AMQP).
	PublishWithOpts(ctx context.Context, msg pkgamqp.Message, sendOpts *amqp.SendOptions) error
}
