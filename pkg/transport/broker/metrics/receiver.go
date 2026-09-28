package metrics

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/transport/broker"
)

// Receiver реализует паттерн проектирования Декоратор (Decorator) поверх абстрактного
// получателя broker.Receiver, обеспечивая сквозной сбор операционных метрик Prometheus.
type Receiver struct {
	receiverName string          // Уникальное имя инстанса получателя для разметки лейблов телеметрии
	receiver     broker.Receiver // Ссылка на оборачиваемый (декорируемый) нижележащий ридер брокера
}

// Гарантируем компиляционную верификацию соответствия интерфейсу broker.Receiver
var _ broker.Receiver = (*Receiver)(nil)

// Receive дожидается нового сообщения. Метод будет обогащен трекингом Latency и инкрементом счетчиков.
func (r Receiver) Receive(ctx context.Context) (broker.Message, error) {
	//TODO implement me
	panic("implement me")
}

// Accept фиксирует успешную обработку. Метод будет обновлять метрику со статусом success.
func (r Receiver) Accept(ctx context.Context, msg broker.Message) error {
	//TODO implement me
	panic("implement me")
}

// Reject отклоняет пакет. Метод будет обновлять счетчик ошибок со статусом fail.
func (r Receiver) Reject(ctx context.Context, msg broker.Message, err error) error {
	//TODO implement me
	panic("implement me")
}

// Release возвращает пакет в очередь. Метод предназначен для трекинга повторных чтений.
func (r Receiver) Release(ctx context.Context, msg broker.Message) error {
	//TODO implement me
	panic("implement me")
}

// Close каскадно закрывает линк и сбрасывает накопленные буферы метрик.
func (r Receiver) Close(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

// GetTargetName возвращает имя топика или очереди, за которыми закреплен декоратор.
func (r Receiver) GetTargetName() string {
	//TODO implement me
	panic("implement me")
}

// Stats агрегирует внутренние показатели рантайма и Prometheus-метрики в единый снимок.
func (r Receiver) Stats() broker.ReceiverStats {
	//TODO implement me
	panic("implement me")
}
