package trace

import (
	"context"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// DataAuditOutbox представляет собой декоратор репозитория для логирования
// и трассировки операций с таблицей Outbox аудита данных.
type DataAuditOutbox struct {
	*trace.CRUD[*extdom.DataAuditOutbox, string]
	next extdom.DataAuditOutboxRepository
}

var _ domain.CRUDRepository[*extdom.DataAuditOutbox, string] = (*DataAuditOutbox)(nil)
var _ extdom.DataAuditOutboxRepository = (*DataAuditOutbox)(nil)
var _ extdom.TailRepository[string] = (*DataAuditOutbox)(nil)

// NewDataAuditOutboxRepository создает новый экземпляр трассирующего декоратора DataAuditOutbox.
func NewDataAuditOutboxRepository(next extdom.DataAuditOutboxRepository) *DataAuditOutbox {
	return &DataAuditOutbox{
		next: next,
		CRUD: trace.NewCRUD("DataAuditOutboxRepository", next),
	}
}

// GetTail возвращает срез идентификаторов старых ("хвостовых") записей outbox, созданных до указанной даты.
func (dao *DataAuditOutbox) GetTail(ctx context.Context, tailDate time.Time) ([]string, error) {
	repoCtx, span := dao.StartSpan(ctx, fmt.Sprintf("%s.GetTail", dao.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.Int64("param.tailDate", tailDate.Unix()),
	)

	res, err := dao.next.GetTail(repoCtx, tailDate)
	if err != nil {
		span.AddEvent("GetTail_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return res, err
}

// CutTail физически удаляет запись из "хвоста" таблицы outbox по её уникальному идентификатору.
func (dao *DataAuditOutbox) CutTail(ctx context.Context, id string) error {
	repoCtx, span := dao.StartSpan(ctx, fmt.Sprintf("%s.CutTail", dao.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("params.id", id),
	)

	err := dao.next.CutTail(repoCtx, id)
	if err != nil {
		span.AddEvent("CutTail_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

// Save сохраняет или обновляет сущность DataAuditOutbox в хранилище в рамках распределенной трассировки.
func (dao *DataAuditOutbox) Save(ctx context.Context, entity *extdom.DataAuditOutbox) error {
	repoCtx, span := dao.StartSpan(ctx, fmt.Sprintf("%s.Save", dao.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("params.entity_id", fmt.Sprintf("%v", entity.GetID())),
	)

	err := dao.next.Save(repoCtx, entity)
	if err != nil {
		span.AddEvent("Save_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

// ChangeStatus атомарно изменяет текущий статус обработки записи outbox (например, с pending на processed).
func (dao *DataAuditOutbox) ChangeStatus(ctx context.Context, id string, status string) error {
	repoCtx, span := dao.StartSpan(ctx, fmt.Sprintf("%s.ChangeStatus", dao.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("params.id", id),
		attribute.String("params.status", status),
	)

	err := dao.next.ChangeStatus(repoCtx, id, status)
	if err != nil {
		span.AddEvent("ChangeStatus_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

// ListPending возвращает список записей outbox, ожидающих отправки, с ограничением по количеству.
func (dao *DataAuditOutbox) ListPending(ctx context.Context, limit int) ([]*extdom.DataAuditOutbox, error) {
	repoCtx, span := dao.StartSpan(ctx, fmt.Sprintf("%s.ListPending", dao.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.Int("param.limit", limit),
	)

	res, err := dao.next.ListPending(repoCtx, limit)
	if err != nil {
		span.AddEvent("ListPending_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return res, err
}

// ListProcessed возвращает список успешно обработанных записей outbox, измененных после указанного времени.
func (dao *DataAuditOutbox) ListProcessed(ctx context.Context, lastUpdatedAt time.Time) ([]*extdom.DataAuditOutbox, error) {
	repoCtx, span := dao.StartSpan(ctx, fmt.Sprintf("%s.ListProcessed", dao.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.Int64("params.last_updated_at", lastUpdatedAt.Unix()),
	)

	res, err := dao.next.ListProcessed(repoCtx, lastUpdatedAt)
	if err != nil {
		span.AddEvent("ListProcessed_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return res, err
}
