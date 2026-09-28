package logger

import (
	"github.com/segmentio/kafka-go"
)

// KafkaLoggerFunc определяет функциональный тип-адаптер для приведения обычных замыканий Go
// к интерфейсу kafka.Logger из библиотеки segmentio/kafka-go.
type KafkaLoggerFunc func(template string, data ...any)

// Printf реализует единственный обязательный контракт интерфейса kafka.Logger,
// прозрачно проксируя форматированные строки логов брокера во внутреннюю функцию-обработчик.
func (f KafkaLoggerFunc) Printf(template string, data ...any) {
	f(template, data...)
}

// KafkaLogger выступает в роли инфраструктурного моста (Adapter), который перехватывает
// внутренние системные логи рантайма Кафки и маршрутизирует их в наш структурированный ZapLogger.
type KafkaLogger struct {
	logger Logger // Ссылка на глобальный абстрактный интерфейс логгера фреймворка
}

// NewKafkaLogger — фабричный конструктор адаптера логирования Кафки.
func NewKafkaLogger(log Logger) *KafkaLogger {
	return &KafkaLogger{
		logger: log,
	}
}

// InfoLogger возвращает адаптированный экземпляр для записи штатных информационных событий брокера
// (например, сообщения о ребалансировке consumer-группы, подключении к координатору или коммите оффсетов).
func (kl *KafkaLogger) InfoLogger() kafka.Logger {
	return KafkaLoggerFunc(func(template string, data ...any) {
		kl.logger.Infof(template, data...)
	})
}

// ErrorLogger возвращает адаптированный экземпляр для логирования сетевых сбоев,
// таймаутов сокетов и ошибок координатора групп брокера (Group Coordinator Not Available).
func (kl *KafkaLogger) ErrorLogger() kafka.Logger {
	return KafkaLoggerFunc(func(template string, data ...any) {
		kl.logger.Errorf(template, data...)
	})
}
