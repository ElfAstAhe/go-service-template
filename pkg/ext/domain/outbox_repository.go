package domain

import (
	"context"
	"time"
)

// OutboxRepository определяет строго типизированный обобщенный доменный интерфейс
// для управления жизненным циклом записей транзакционного журнала Outbox (Transactional Outbox Pattern).
//
// Абстрагирует низкоуровневую реализацию работы с хранилищем (SQL/NoSQL) от прикладного слоя репозиториев,
// обеспечивая сохранение слепков аудит-логов в рамках единой ACID-транзакции бизнес-логики.
type OutboxRepository[E Outbox[ID], ID comparable] interface {
	// Save осуществляет персистентную запись новой outbox-сущности в локальное хранилище.
	// Принимает контекст выполнения, разделяемый с родительской транзакцией Use Case, и возвращает созданную сущность либо ошибку.
	Save(context.Context, E) error
	ChangeStatus(ctx context.Context, id ID, status string) error

	// ListPending извлекает из хранилища пачку неотправленных outbox-сущностей заданного лимита.
	ListPending(ctx context.Context, limit int) ([]E, error)
	ListProcessed(ctx context.Context, lastUpdatedAt time.Time) ([]E, error)
}
