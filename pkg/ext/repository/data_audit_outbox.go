package postgres

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/repository"
)

const (
	sourceLabelListPending   string = "ListPending"
	sourceLabelListProcessed string = "ListProcessed"
	sourceLabelGetTail       string = "GetTail"
)

type DataAuditOutboxRepository struct {
	*repository.CRUD[*extdom.DataAuditOutbox, string]
}

var _ domain.CRUDRepository[*extdom.DataAuditOutbox, string] = (*DataAuditOutboxRepository)(nil)
var _ extdom.OutboxRepository[*extdom.DataAuditOutbox, string] = (*DataAuditOutboxRepository)(nil)
var _ extdom.TailRepository[string] = (*DataAuditOutboxRepository)(nil)
var _ extdom.DataAuditOutboxRepository = (*DataAuditOutboxRepository)(nil)

func NewDataAuditOutboxRepository(
	querier db.Querier,
	decipher db.ErrorDecipher,
) (*DataAuditOutboxRepository, error) {
	// instance
	res := &DataAuditOutboxRepository{}
	// sql

	// callbacks
	// embed
	// setup

	return res, nil
}

func (dao *DataAuditOutboxRepository) GetTail(ctx context.Context, tailDate time.Time) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *DataAuditOutboxRepository) Save(ctx context.Context, entity *extdom.DataAuditOutbox) error {
	_, err := dao.Create(ctx, entity)

	return err
}

func (dao *DataAuditOutboxRepository) ChangeStatus(ctx context.Context, id string, status string) error {
	//TODO implement me
	panic("implement me")
}

func (dao *DataAuditOutboxRepository) ListPending(ctx context.Context, limit int) ([]*extdom.DataAuditOutbox, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *DataAuditOutboxRepository) ListProcessed(ctx context.Context, lastUpdatedAt time.Time) ([]*extdom.DataAuditOutbox, error) {
	//TODO implement me
	panic("implement me")
}
