package azure

import (
	"strings"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
)

const (
	DefaultReceiverConnectTimeout  time.Duration = 5 * time.Second
	DefaultReceiverShutdownTimeout time.Duration = 5 * time.Second
	DefaultReceiverLinkCredit      int32         = 100 // Наш золотой дефолт для Flow Control
)

type ReceiverOption func(*ReceiverOptions)

type ReceiverOptions struct {
	Connector       pkgamqp.Connector[*amqp.Session] // Ссылка на наш общий дженерик-коннектор
	TargetName      string                           // Имя конкретной очереди/топика для сингл-ресивера
	ReceiverOpts    *amqp.ReceiverOptions            // Кастомные опции Azure AMQP
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	Logger          logger.Logger
	LinkCredit      int32 // Инкапсулированная настройка кредитов
}

func NewReceiverOptions() *ReceiverOptions {
	return &ReceiverOptions{
		ConnectTimeout:  DefaultReceiverConnectTimeout,
		ShutdownTimeout: DefaultReceiverShutdownTimeout,
		LinkCredit:      DefaultReceiverLinkCredit,
	}
}

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
	if ro.LinkCredit <= 0 {
		return errs.NewTlCommonError("Validate", "link credit must be greater than zero", nil)
	}

	return nil
}

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
