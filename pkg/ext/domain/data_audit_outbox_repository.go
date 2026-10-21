package domain

import (
	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// DataAuditOutboxRepository определяет специализированный интерфейс доменного уровня
// для управления персистентным состоянием записей транзакционного журнала Outbox.
type DataAuditOutboxRepository interface {
	pkgdom.CRUDRepository[*DataAuditOutbox, string]
	TailRepository[string]
	OutboxRepository[*DataAuditOutbox, string]
}
