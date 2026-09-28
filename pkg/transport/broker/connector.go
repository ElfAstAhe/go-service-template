package broker

import (
	"context"
)

// Connector определяет унифицированный, строго типизированный (Generic) контракт
// для управления физическими подключениями к брокерам очередей и внешним СУБД (Standardized Connection Lifecycle).
//
// Инкапсулирует рутину мягкого высвобождения ресурсов, ленивого извлечения сокетов и проактивной
// рантайм-инвалидации битых сетевых сессий при обнаружении инфраструктурных сбоев.
type Connector[Connection any] interface {
	// Close осуществляет каскадное закрытие активного пула соединений в рамках лимитов контекста ctx.
	Close(ctx context.Context) error

	// GetConnection возвращает валидный, готовый к работе сетевой сокет/клиент типа Connection.
	GetConnection(ctx context.Context) (Connection, error)

	// Invalidate помечает текущую сессию как битую (Stale/Broken) при перехвате сетевых ошибок err,
	// инициируя автоматический фоновый реконнект (failover) при последующем вызове GetConnection.
	Invalidate(err error)
}
