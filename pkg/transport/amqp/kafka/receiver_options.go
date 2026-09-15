package kafka

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/segmentio/kafka-go"
)

const (
	DefaultReceiverConnectTimeout  time.Duration = 5 * time.Second
	DefaultReceiverShutdownTimeout time.Duration = 5 * time.Second
	DefaultReceiverMinBytes        int           = 10e3 // 10 KB — минимальный пакет для вычитки
	DefaultReceiverMaxBytes        int           = 10e6 // 10 MB — максимальный пакет на одну итерацию
	DefaultReceiverMaxWait         time.Duration = 1 * time.Second
)

type ReceiverOption func(*ReceiverOptions)

type ReceiverOptions struct {
	Connector       pkgamqp.Connector[*kafka.Client] // Ссылка на наш общий дженерик-коннектор Kafka
	TargetName      string                           // Имя конкретного топика (Topic)
	GroupID         string                           // Идентификатор Consumer Group (Критично для Kafka)
	KafkaReaderOpts *kafka.ReaderConfig              // Кастомные сырые опции segmentio/kafka-go
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	Logger          logger.Logger
	MinBytes        int
	MaxBytes        int
	MaxWait         time.Duration
}

func NewReceiverOptions() *ReceiverOptions {
	return &ReceiverOptions{
		ConnectTimeout:  DefaultReceiverConnectTimeout,
		ShutdownTimeout: DefaultReceiverShutdownTimeout,
		MinBytes:        DefaultReceiverMinBytes,
		MaxBytes:        DefaultReceiverMaxBytes,
		MaxWait:         DefaultReceiverMaxWait,
	}
}

func (ro *ReceiverOptions) Validate() error {
	if ro.Connector == nil {
		return errs.NewTlCommonError("Validate", "connector is required and cannot be nil", nil)
	}
	if strings.TrimSpace(ro.TargetName) == "" {
		return errs.NewTlCommonError("Validate", "target name (topic) cannot be empty", nil)
	}
	if strings.TrimSpace(ro.GroupID) == "" {
		return errs.NewTlCommonError("Validate", "group id (consumer group) cannot be empty for kafka", nil)
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
	if ro.MinBytes <= 0 || ro.MaxBytes <= 0 {
		ro.MinBytes = DefaultReceiverMinBytes
		ro.MaxBytes = DefaultReceiverMaxBytes
	}

	return nil
}

func WithReceiverConnector(connector pkgamqp.Connector[*kafka.Client]) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.Connector = connector
	}
}

func WithReceiverTargetName(targetName string) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.TargetName = targetName
	}
}

func WithReceiverGroupID(groupID string) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.GroupID = groupID
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

func WithReceiverFetchBounds(minBytes, maxBytes int, maxWait time.Duration) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.MinBytes = minBytes
		ro.MaxBytes = maxBytes
		ro.MaxWait = maxWait
	}
}

func WithKafkaReaderOpts(readerOpts *kafka.ReaderConfig) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.KafkaReaderOpts = readerOpts
	}
}
