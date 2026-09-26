package kafka

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
)

const sysKafkaMsgKey = "_sys_kafka_orig_message"

// Message реализует ваш интерфейс pkgamqp.Message
type Message struct {
	TargetName string
	Payload    []byte
	Props      map[string]any
}

var _ pkgamqp.Message = (*Message)(nil)

// NewMessage собирает наш конверт из сырого сообщения библиотеки kafka-go.
// Заголовки (Headers) из Kafka автоматически маппятся в общую карту Props.
func NewMessage(msg kafka.Message) *Message {
	props := make(map[string]any)

	// Переносим заголовки Kafka в универсальную карту свойств
	for _, h := range msg.Headers {
		props[h.Key] = string(h.Value)
	}

	// Сохраняем оригинальный пакет для метода ExtractOriginalMessage.
	// Сохраняем как указатель, чтобы избежать лишнего копирования структуры в памяти.
	props[sysKafkaMsgKey] = &msg

	return &Message{
		TargetName: msg.Topic,
		Payload:    msg.Value,
		Props:      props,
	}
}

// GetTargetName возвращает имя топика Kafka
func (m *Message) GetTargetName() string {
	return m.TargetName
}

// GetPayload возвращает тело (Value) сообщения Kafka
func (m *Message) GetPayload() []byte {
	return m.Payload
}

// GetProperties возвращает заголовки сообщения в виде плоской карты
func (m *Message) GetProperties() map[string]any {
	return m.Props
}

// ExtractOriginalMessage безопасно извлекает указатель на сырую структуру *kafka.Message.
// Это полезно, если на уровне обработчика потребуется сделать ручной Commit (узнать Partition, Offset и т.д.).
func (m *Message) ExtractOriginalMessage() (any, error) {
	if utils.IsNil(m.Props) {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "properties map is nil", nil)
	}
	raw, ok := m.Props[sysKafkaMsgKey]
	if !ok {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "metadata missing", nil)
	}

	return raw, nil
}
