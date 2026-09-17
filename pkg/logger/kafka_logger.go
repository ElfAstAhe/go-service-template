package logger

import (
	"github.com/segmentio/kafka-go"
)

type KafkaLoggerFunc func(template string, data ...any)

func (f KafkaLoggerFunc) Printf(template string, data ...any) {
	f(template, data...)
}

type KafkaLogger struct {
	logger Logger
}

func NewKafkaLogger(log Logger) *KafkaLogger {
	return &KafkaLogger{
		logger: log,
	}
}

func (kl *KafkaLogger) InfoLogger() kafka.Logger {
	return KafkaLoggerFunc(func(template string, data ...any) {
		kl.logger.Infof(template, data...)
	})
}

func (kl *KafkaLogger) ErrorLogger() kafka.Logger {
	return KafkaLoggerFunc(func(template string, data ...any) {
		kl.logger.Errorf(template, data...)
	})
}
