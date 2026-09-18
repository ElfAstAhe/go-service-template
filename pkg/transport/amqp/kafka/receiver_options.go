package kafka

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/segmentio/kafka-go"
)

const (
	DefaultReceiverConnectTimeout  time.Duration = 5 * time.Second
	DefaultReceiverShutdownTimeout time.Duration = 5 * time.Second
	DefaultReceiverMinBytes        int           = 10e3 // 10 KB
	DefaultReceiverMaxBytes        int           = 10e6 // 10 MB
	DefaultReceiverMaxWait         time.Duration = 1 * time.Second
)

type ReceiverOption func(*ReceiverOptions)

type ReceiverOptions struct {
	Brokers         []string            // Заменили Connector на прямой список хостов брокеров
	TargetName      string              // Имя топика (Topic)
	GroupID         string              // Идентификатор Consumer Group
	KafkaReaderOpts *kafka.ReaderConfig // Дополнительные кастомные опции
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
	if len(ro.Brokers) == 0 {
		return errs.NewTlCommonError("Validate", "at least one broker address is required", nil)
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

func WithReceiverBrokers(brokers []string) ReceiverOption {
	return func(ro *ReceiverOptions) {
		ro.Brokers = brokers
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
