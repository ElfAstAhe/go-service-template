package amqp

import (
	"context"
)

// ClientSingleSender предназначен для работы по схеме 1 к 1.
// Он жестко завязан на одну конкретную очередь/топик, переданную в конструктор.
type ClientSingleSender[SendOpts any] interface {
	Publish(ctx context.Context, msg *Message, opts SendOpts) error
	Close(ctx context.Context) error

	GetTargetName() string
}

// ClientMultiSender предназначен для динамической маршрутизации (1 ко многим).
// Идеально разрешает проблему протекающих абстракций с помощью двух типов опций:
// SenderOpts — для ленивого создания/настройки линка-отправителя под конкретный targetName
// SendOpts — для параметров отправки конкретного сообщения (кадра)
type ClientMultiSender[SenderOpts any, SendOpts any] interface {
	Publish(ctx context.Context, targetName string, senderOpts SenderOpts, msg *Message, sendOpts SendOpts) error
	Close(ctx context.Context) error

	GetTargetNames() []string
}
