package metrics

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository/metrics"
)

// DataAuditOutbox декоратор репозитория domain.DataAuditOutbox для сбора метрик
type DataAuditOutbox struct {
	*metrics.CRUD[*extdom.DataAuditOutbox, string]
	next extdom.DataAuditOutboxRepository
}

var _ domain.CRUDRepository[*extdom.DataAuditOutbox, string] = (*DataAuditOutbox)(nil)
var _ extdom.DataAuditOutboxRepository = (*DataAuditOutbox)(nil)
var _ extdom.TailRepository[string] = (*DataAuditOutbox)(nil)

// NewDataAuditOutboxRepository фабричный метод создания экземпляра декоратора metrics.DataAuditOutbox репозитория domain.DataAuditOutboxRepository
func NewDataAuditOutboxRepository(next extdom.DataAuditOutboxRepository) *DataAuditOutbox {
	return &DataAuditOutbox{
		next: next,
		CRUD: metrics.NewCRUD("DataAuditOutboxRepository", next),
	}
}

// GetTail получение списка на удаление, реализация интерфейса domain.TailRepository
func (dao *DataAuditOutbox) GetTail(ctx context.Context, tailDate time.Time) (res []string, err error) {
	defer func(startTime time.Time) {
		metrics.ObserveRepositoryOp(dao.GetRepositoryName(), "GetTail", err, startTime)
	}(time.Now())

	return dao.next.GetTail(ctx, tailDate)
}

// CutTail удаление одного элемента, реализация domain.TailRepository
func (dao *DataAuditOutbox) CutTail(ctx context.Context, id string) (err error) {
	defer func(startTime time.Time) {
		metrics.ObserveRepositoryOp(dao.GetRepositoryName(), "CutTail", err, startTime)
	}(time.Now())

	return dao.next.CutTail(ctx, id)
}

// Save сохранение элемента domain.DataAuditOutbox, реализация domain.OutboxRepository
func (dao *DataAuditOutbox) Save(ctx context.Context, entity *extdom.DataAuditOutbox) (err error) {
	defer func(startTime time.Time) {
		metrics.ObserveRepositoryOp(dao.GetRepositoryName(), "Save", err, startTime)
	}(time.Now())

	return dao.next.Save(ctx, entity)
}

// ChangeStatus меняет статус элемента domain.DataAuditOutbox, реализация domain.OutboxRepository
func (dao *DataAuditOutbox) ChangeStatus(ctx context.Context, id string, status string) (err error) {
	defer func(startTime time.Time) {
		metrics.ObserveRepositoryOp(dao.GetRepositoryName(), "ChangeStatus", err, startTime)
	}(time.Now())

	return dao.next.ChangeStatus(ctx, id, status)
}

// ListPending возвращает список необработанных элементов domain.DataAuditOutbox, реализация domain.OutboxRepository
func (dao *DataAuditOutbox) ListPending(ctx context.Context, limit int) (res []*extdom.DataAuditOutbox, err error) {
	defer func(startTime time.Time) {
		metrics.ObserveRepositoryOp(dao.GetRepositoryName(), "ListPending", err, startTime)
	}(time.Now())

	return dao.next.ListPending(ctx, limit)
}

// ListProcessed возвращает список обработанных элементов domain.DataAuditOutbox, реализация domain.OutboxRepository
func (dao *DataAuditOutbox) ListProcessed(ctx context.Context, lastUpdatedAt time.Time) (res []*extdom.DataAuditOutbox, err error) {
	defer func(startTime time.Time) {
		metrics.ObserveRepositoryOp(dao.GetRepositoryName(), "ListProcessed", err, startTime)
	}(time.Now())

	return dao.next.ListProcessed(ctx, lastUpdatedAt)
}
