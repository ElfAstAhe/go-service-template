package pubsub

import (
	"context"
)

// Publisher описывает контракт издателя событий (Subject / Observable).
// Управляет динамическим реестром подписчиков и координирует веерную рассылку (Fan-Out) уведомлений.
type Publisher[T any] interface {
	// Register добавляет нового наблюдателя в текущий пул активных подписчиков.
	Register(observer Observer[T])

	// Unregister атомарно исключает наблюдателя из реестра рассылки (например, при Graceful Shutdown компонента).
	Unregister(observer Observer[T])

	// Notify запускает цикл последовательного или параллельного оповещения всех зарегистрированных наблюдателей.
	Notify(ctx context.Context, event T)
}

// Observer описывает контракт конкретного наблюдателя (Subscriber / Listener).
// Реализует целевую прикладную реакцию на наступившее в системе событие типа T.
type Observer[T any] interface {
	// GetName возвращает уникальный строковый идентификатор наблюдателя.
	// Используется для трассировки цепочки вызовов, сквозного логирования и разделения метрик Prometheus.
	GetName() string

	// OnNotify вызывается издателем при наступлении события.
	// Протаскивание context.Context гарантирует строгий контроль таймаутов и сквозных TraceID.
	// Возврат ошибки позволяет издателю контролировать успешность доставки уведомления.
	OnNotify(ctx context.Context, event T) error
}
