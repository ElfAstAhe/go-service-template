package postgres

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/db/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	apppg "github.com/ElfAstAhe/go-service-template/pkg/ext/repository/postgres"
	pkgrepo "github.com/ElfAstAhe/go-service-template/pkg/repository/v2"
)

const (
	statementListPending   string = "ListPending"
	statementListProcessed string = "ListProcessed"
	statementChangeStatus  string = "ChangeStatus"
	statementGetTail       string = "GetTail"
)

// DataAuditOutbox представляет собой PostgreSQL-репозиторий для управления
// сообщениями аудита данных в паттерне Outbox. Он расширяет базовый CRUD-репозиторий.
type DataAuditOutbox struct {
	*pkgrepo.CRUD[*extdom.DataAuditOutbox, string]
}

var _ domain.CRUDRepository[*extdom.DataAuditOutbox, string] = (*DataAuditOutbox)(nil)
var _ extdom.OutboxRepository[*extdom.DataAuditOutbox, string] = (*DataAuditOutbox)(nil)
var _ extdom.TailRepository[string] = (*DataAuditOutbox)(nil)
var _ extdom.DataAuditOutboxRepository = (*DataAuditOutbox)(nil)

// NewDataAuditOutbox создает и настраивает новый экземпляр DataAuditOutbox репозитория
// с внедренным базовым CRUD-функционалом, хуками и SQL-стейтментами.
func NewDataAuditOutbox(
	executor db.Executor,
	errDecipher db.ErrorDecipher,
) (*DataAuditOutbox, error) {
	// instance
	res := &DataAuditOutbox{}
	// embed
	embed, err := pkgrepo.NewCRUD[*extdom.DataAuditOutbox, string](
		pkgrepo.WithExecutor[*extdom.DataAuditOutbox, string](executor),
		pkgrepo.WithErrorDecipher[*extdom.DataAuditOutbox, string](errDecipher),
		pkgrepo.WithHooks[*extdom.DataAuditOutbox, string](
			pkgrepo.WithHookNewEntityFactory[*extdom.DataAuditOutbox, string](extdom.NewEmptyDataAuditOutbox),
			pkgrepo.WithHookCreator[*extdom.DataAuditOutbox, string](res.creator),
			pkgrepo.WithHookChanger[*extdom.DataAuditOutbox, string](res.changer),
		),
		pkgrepo.WithEntityInfo[*extdom.DataAuditOutbox, string]("data_audit_outbox", "DataAuditOutbox"),
		pkgrepo.WithStatementsBuilder[*extdom.DataAuditOutbox, string](postgres.DriverName, res.pgStatementsBuilder),
		pkgrepo.WithStatementsBuilder[*extdom.DataAuditOutbox, string]("pgx", res.pgStatementsBuilder),
		pkgrepo.WithStatementsBuilder[*extdom.DataAuditOutbox, string]("postgres", res.pgStatementsBuilder),
	)
	if err != nil {
		return nil, errs.NewDalError("NewDataAuditOutbox", "create embed crud failed", err)
	}
	// setup
	res.CRUD = embed

	return res, nil
}

// GetTail возвращает список идентификаторов устаревших записей, время изменения которых
// меньше указанной даты. Реализует интерфейс extdom.TailRepository.
func (dao *DataAuditOutbox) GetTail(ctx context.Context, tailDate time.Time) ([]string, error) {
	list, err := dao.GetHelper().List(ctx, statementGetTail, tailDate)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.GetTail", "get tail ids failed", err)
	}

	return domain.MapEntitiesToIDs(list), nil
}

// Save выполняет сохранение (создание) новой записи outbox в базе данных.
// Реализует интерфейс extdom.OutboxRepository.
func (dao *DataAuditOutbox) Save(ctx context.Context, entity *extdom.DataAuditOutbox) error {
	_, err := dao.Create(ctx, entity)

	return err
}

// ChangeStatus обновляет статус конкретной записи outbox по её идентификатору.
// Реализует интерфейс extdom.OutboxRepository.
func (dao *DataAuditOutbox) ChangeStatus(ctx context.Context, id string, status string) error {
	if err := dao.GetHelper().ExecNoCheck(ctx, statementChangeStatus, id, status); err != nil {
		return errs.NewDalError("DataAuditOutbox.ChangeStatus", "change status failed", err)
	}

	return nil
}

// ListPending извлекает список необработанных записей outbox с ограничением по количеству.
// Реализует интерфейс extdom.OutboxRepository.
func (dao *DataAuditOutbox) ListPending(ctx context.Context, limit int) ([]*extdom.DataAuditOutbox, error) {
	return dao.GetHelper().List(ctx, statementListPending, limit)
}

// ListProcessed извлекает список успешно обработанных записей, обновленных после указанного времени.
// Реализует интерфейс extdom.OutboxRepository.
func (dao *DataAuditOutbox) ListProcessed(ctx context.Context, lastUpdatedAt time.Time) ([]*extdom.DataAuditOutbox, error) {
	return dao.GetHelper().List(ctx, statementListProcessed, lastUpdatedAt)
}

// creator является внутренним хуком для выполнения SQL-запроса на добавление новой строки в базу данных.
func (dao *DataAuditOutbox) creator(ctx context.Context, querier db.Querier, entity *extdom.DataAuditOutbox, params ...any) (pkgrepo.Scannable, error) {
	query, err := dao.GetHelper().GetStatements().GetSQL(pkgrepo.StatementCreate)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.Creator", "create statement not found", err)
	}

	return querier.QueryRowContext(ctx, query, params...), nil
}

// changer является внутренним хуком для выполнения SQL-запроса на изменение существующей строки в базе данных.
func (dao *DataAuditOutbox) changer(ctx context.Context, querier db.Querier, entity *extdom.DataAuditOutbox, params ...any) (pkgrepo.Scannable, error) {
	query, err := dao.GetHelper().GetStatements().GetSQL(pkgrepo.StatementChange)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.Creator", "change statement not found", err)
	}

	return querier.QueryRowContext(ctx, query, params...), nil
}

// pgStatementsBuilder инициализирует и связывает SQL-запросы PostgreSQL с внутренними идентификаторами стейтментов.
func (dao *DataAuditOutbox) pgStatementsBuilder() *pkgrepo.Statements[*extdom.DataAuditOutbox, string] {
	return pkgrepo.NewStatements[*extdom.DataAuditOutbox, string]().
		WithStatement(pkgrepo.StatementFind, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxFind, dao.mapDefault)).
		WithStatement(pkgrepo.StatementList, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxList, dao.mapDefault)).
		WithStatement(pkgrepo.StatementCreate, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxCreate, dao.mapDefault)).
		WithStatement(pkgrepo.StatementChange, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxChange, dao.mapDefault)).
		WithStatement(pkgrepo.StatementDelete, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxDelete, dao.mapDefault)).
		WithStatement(statementGetTail, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxGetTail, dao.mapTail)).
		WithStatement(statementListPending, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxListPending, dao.mapDefault)).
		WithStatement(statementListProcessed, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxListProcessed, dao.mapDefault)).
		WithStatement(statementChangeStatus, pkgrepo.NewStatement[*extdom.DataAuditOutbox, string](apppg.SqlDataAuditOutboxChangeStatus, nil))
}

// mapDefault преобразует сырые строки базы данных (Scannable) в полноценный объект DataAuditOutbox со всеми заполненными полями.
func (dao *DataAuditOutbox) mapDefault(ctx context.Context, scannable pkgrepo.Scannable, params ...any) (*extdom.DataAuditOutbox, error) {
	res := dao.GetHelper().GetHooks().NewEntityFactory()
	err := scannable.Scan(
		&res.ID,
		&res.InstanceID,
		&res.ContentType,
		&res.Payload,
		&res.Status,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.mapDefault", "default mapper failed", err)
	}

	return res, nil
}

// mapTail выполняет частичный маппинг строки базы данных, считывая исключительно поле идентификатора (ID).
func (dao *DataAuditOutbox) mapTail(ctx context.Context, scannable pkgrepo.Scannable, params ...any) (*extdom.DataAuditOutbox, error) {
	res := dao.GetHelper().GetHooks().NewEntityFactory()
	err := scannable.Scan(
		&res.ID,
	)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.mapTail", "tail mapper failed", err)
	}

	return res, nil
}
