package metrics

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
)

type Receiver struct {
}

func (r Receiver) Receive(ctx context.Context) (broker.Message, error) {
	//TODO implement me
	panic("implement me")
}

func (r Receiver) Accept(ctx context.Context, msg broker.Message) error {
	//TODO implement me
	panic("implement me")
}

func (r Receiver) Reject(ctx context.Context, msg broker.Message, err error) error {
	//TODO implement me
	panic("implement me")
}

func (r Receiver) Release(ctx context.Context, msg broker.Message) error {
	//TODO implement me
	panic("implement me")
}

func (r Receiver) Close(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (r Receiver) GetTargetName() string {
	//TODO implement me
	panic("implement me")
}

func (r Receiver) Stats() broker.ReceiverStats {
	//TODO implement me
	panic("implement me")
}

var _ broker.Receiver = (*Receiver)(nil)
