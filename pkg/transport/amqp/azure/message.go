package azure

import (
	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
)

const sysMsgKey = "_sys_amqp_orig_azure_message"

type Message struct {
	TargetName string
	Header     *amqp.MessageHeader
	Payload    []byte
	Props      map[string]any
}

var _ pkgamqp.Message = (*Message)(nil)

func NewMessage(payload []byte, props map[string]any) *Message {
	return &Message{
		Payload: payload,
		Props:   props,
	}
}

func (m *Message) GetTargetName() string {
	return m.TargetName
}

func (m *Message) GetPayload() []byte {
	return m.Payload
}

func (m *Message) GetProperties() map[string]any {
	return m.Props
}

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
