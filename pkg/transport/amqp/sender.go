package amqp

import (
	"context"
)

// Sender предназначен для работы по схеме 1 к 1.
// Он жестко завязан на одну конкретную очередь/топик, переданную в конструктор.
//
//	SendOpts - для параметров отправки конкретного сообщения
//	MsgHeader - заголовок сообщения
type Sender[SendOpts any, MsgHeader any] interface {
	Publish(ctx context.Context, msg *Message[MsgHeader], opts SendOpts) error
	Close(ctx context.Context) error

	GetTargetName() string
}
