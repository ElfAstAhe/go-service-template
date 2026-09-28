package azure

import (
	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
)

const sysMsgKey = "_sys_amqp_orig_azure_message"

// Message реализует интерфейс pkgamqp.Message, оборачивая параметры пакета AMQP 1.0.
// Служит независимым контейнером полезной нагрузки и метаданных для UseCase-слоя платформы.
type Message struct {
	TargetName string              // Имя целевой очереди или топика брокера
	Header     *amqp.MessageHeader // Низкоуровневые системные заголовки пакета спецификации AMQP
	Payload    []byte              // Бинарное тело (бизнес-данные) сообщения
	Props      map[string]any      // Карта пользовательских атрибутов и служебных метаданных
}

// Гарантируем компиляционную верификацию соответствия интерфейсу pkgamqp.Message
var _ broker.Message = (*Message)(nil)

// NewMessage — фабричный конструктор прикладного сообщения для публикации в брокер очередей.
func NewMessage(payload []byte, props map[string]any) *Message {
	return &Message{
		Payload: payload,
		Props:   props,
	}
}

// GetTargetName возвращает строковое имя очереди или топика назначения.
func (m *Message) GetTargetName() string {
	return m.TargetName
}

// GetPayload возвращает сырой массив байт полезной нагрузки сообщения.
func (m *Message) GetPayload() []byte {
	return m.Payload
}

// GetProperties возвращает карту метаданных и заголовков текущего пакета.
func (m *Message) GetProperties() map[string]any {
	return m.Props
}

// ExtractOriginalMessage безопасно извлекает нативный объект *amqp.Message из карты свойств Props.
// Используется воркерами для ручного контроля жизненного цикла сетевых пакетов брокера.
func (m *Message) ExtractOriginalMessage() (any, error) {
	if !(len(m.Props) > 0) {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "envelope props empty", nil)
	}
	raw, exists := m.Props[sysMsgKey]
	if !exists {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "original message not exists", nil)
	}
	res, ok := (raw).(*amqp.Message)
	if !ok {
		return nil, errs.NewTlCommonError("ExtractOriginalMessage", "invalid underlying packet structure type", nil)
	}

	return res, nil
}
