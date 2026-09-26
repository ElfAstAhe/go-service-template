package amqp

import (
	"strings"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
)

// Константы рантайм-дефолтов для внутренней защиты получателя Azure Service Bus.
const (
	DefaultReceiverConnectTimeout  time.Duration = 5 * time.Second
	DefaultReceiverShutdownTimeout time.Duration = 5 * time.Second
	DefaultReceiverLinkCredit      int32         = 100 // Наш золотой дефолт для Flow Control
)

// ReceiverOption определяет функциональный тип для конфигурации опций получателя (Fluent API).
type ReceiverOption func(*ReceiverOptions)

// ReceiverOptions инкапсулирует все параметры рантайма, необходимые для безопасной сборки и работы Receiver.
type ReceiverOptions struct {
	// Connector представляет собой ссылку на глобальный дженерик-коннектор для управления AMQP-сессиями.
	Connector pkgamqp.Connector[*amqp.Session] // Ссылка на наш общий дженерик-коннектор
	// TargetName определяет точное имя целевой очереди (Queue) или подписки на топик (Subscription).
	TargetName string // Имя конкретной очереди/топика для сингл-ресивера
	// ReceiverOpts содержит специфичные низкоуровневые настройки линка вычитки библиотеки Azure go-amqp.
	ReceiverOpts *amqp.ReceiverOptions // Кастомные опции Azure AMQP
	// ReceiveOpts инкапсулирует дефолтные параметры для единичных вызовов Fetch (используется плоским Receive).
	ReceiveOpts     *amqp.ReceiveOptions
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	Logger          logger.Logger
	// LinkCredit задает количество сообщений, которое брокер может отправить в буфер линка без подтверждения (Credits).
	LinkCredit int32 // Инкапсулированная настройка кредитов
}

// NewReceiverOptions инициализирует структуру опций, сразу наполняя её безопасными базовыми таймаутами и кредитами.
func NewReceiverOptions() *ReceiverOptions {
	return &ReceiverOptions{
		ConnectTimeout:  DefaultReceiverConnectTimeout,
		ShutdownTimeout: DefaultReceiverShutdownTimeout,
		LinkCredit:      DefaultReceiverLinkCredit,
	}
}

// Validate выполняет строгую проверку входящих параметров рантайма перед созданием Receiver.
// Защищает приложение от паник при отсутствии обязательных инфраструктурных зависимостей и логов.
func (ro *ReceiverOptions) Validate() error {
	if ro.Connector == nil {
		return errs.NewTlCommonError("Validate", "connector is required and cannot be nil", nil)
	}
	if strings.TrimSpace(ro.TargetName) == "" {
		return errs.NewTlCommonError("Validate", "target name (queue/topic) cannot be empty", nil)
	}
	if ro.Logger == nil {
		return errs.NewTlCommonError("Validate", "logger is nil", nil)
	}
	if ro.ConnectTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "connection timeout is invalid", nil)
	}
	if ro.ShutdownTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "shutdown timeout is invalid", nil)
	}
	// Золотое правило кредитования AMQP: значение Flow Control должно позволять буферизацию
	if ro.LinkCredit <= 0 {
		return errs.NewTlCommonError("Validate", "link credit must be greater than zero", nil)
	}

	return nil
}

// ====================================================================
// Fluent API методы конфигурации
// ====================================================================

func WithReceiverConnector(connector pkgamqp.Connector[*amqp.Session]) ReceiverOption {
	return func(cro *ReceiverOptions) {
		cro.Connector = connector
	}
}

func WithReceiverTargetName(targetName string) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.TargetName = targetName
	}
}

func WithReceiverConnectTimeout(timeout time.Duration) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.ConnectTimeout = timeout
	}
}

func WithReceiverShutdownTimeout(timeout time.Duration) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.ShutdownTimeout = timeout
	}
}

func WithReceiverLogger(log logger.Logger) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.Logger = log
	}
}

func WithReceiverLinkCredit(credit int32) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.LinkCredit = credit
	}
}

func WithReceiverOpts(receiverOpts *amqp.ReceiverOptions) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.ReceiverOpts = receiverOpts
	}
}

func WithReceiverReceiveOpts(receiveOpts *amqp.ReceiveOptions) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.ReceiveOpts = receiveOpts
	}
}
