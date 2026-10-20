package domain

import (
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// Outbox определяет обобщенный платформенный контракт сущности типа outbox (Транзакционный лог отправки).
//
// Использует дженерик-ограничение [ID comparable] для поддержки любых типов идентификаторов
// (UUID, int64, string) в дочерних микросервисах. Наследует базовые свойства domain.Entity[ID],
// фиксируя принадлежность к слою чистых сущностей по канонам Domain-Driven Design (DDD).
type Outbox[ID comparable] interface {
	domain.Entity[ID]

	// GetKey возвращает уникальный бизнес-ключ сообщения (например, InstanceID или PartitionKey),
	// используемый брокерами данных для строгого распределения событий по партициям.
	GetKey() string

	// GetContentType возвращает тип контента payload (например application/json),
	// необходимый транспортному слою для правильной десериализации тела сообщения.
	GetContentType() string

	// GetPayload вовзращает сырое содержимое бизнес-события на отправку в виде среза байт.
	GetPayload() []byte

	// GetStatus возвращает текущий статус отправки записи (например: PENDING, SUCCESS, FAILED),
	// константы OutboxStatusPending, OutboxStatusSuccess, OutboxStatusFailed.
	GetStatus() string
}
