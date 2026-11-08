package postgres

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	pg "github.com/ElfAstAhe/go-service-template/pkg/db/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	extpg "github.com/ElfAstAhe/go-service-template/pkg/ext/repository/postgres"
	repo "github.com/ElfAstAhe/go-service-template/pkg/repository/v2"
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
	*repo.CRUD[*extdom.DataAuditOutbox, string]
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
	embed, err := repo.NewCRUD[*extdom.DataAuditOutbox, string](
		repo.WithExecutor[*extdom.DataAuditOutbox, string](executor),
		repo.WithErrorDecipher[*extdom.DataAuditOutbox, string](errDecipher),
		repo.WithHooks[*extdom.DataAuditOutbox, string](
			repo.WithHookNewEntityFactory[*extdom.DataAuditOutbox, string](extdom.NewEmptyDataAuditOutbox),
			repo.WithHookCreator[*extdom.DataAuditOutbox, string](res.creator),
			repo.WithHookChanger[*extdom.DataAuditOutbox, string](res.changer),
		),
		repo.WithEntityInfo[*extdom.DataAuditOutbox, string]("data_audit_outbox", "DataAuditOutbox"),
		repo.WithStatementsBuilder[*extdom.DataAuditOutbox, string](pg.DriverName, res.pgStatementsBuilder),
		repo.WithStatementsBuilder[*extdom.DataAuditOutbox, string]("pgx", res.pgStatementsBuilder),
		repo.WithStatementsBuilder[*extdom.DataAuditOutbox, string]("postgres", res.pgStatementsBuilder),
	)
	if err != nil {
		return nil, errs.NewDalError("NewDataAuditOutbox", "create embed crud failed", err)
	}
	// setup
	res.CRUD = embed

	return res, nil
}

// GetTail возвращает список идентификаторов устаревших записей, время изменения которых
// меньше указанной даты. Реализует интерфейс domain.TailRepository.
func (dao *DataAuditOutbox) GetTail(ctx context.Context, tailDate time.Time) ([]string, error) {
	list, err := dao.GetHelper().List(ctx, statementGetTail, tailDate)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.GetTail", "get tail ids failed", err)
	}

	return domain.MapEntitiesToIDs(list), nil
}

// CutTail точечно удаляет элемент хвоста. Реализует интерфейс domain.TailRepository
func (dao *DataAuditOutbox) CutTail(ctx context.Context, id string) error {
	return dao.Delete(ctx, id)
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
func (dao *DataAuditOutbox) creator(ctx context.Context, querier db.Querier, entity *extdom.DataAuditOutbox, params ...any) (repo.Scannable, error) {
	query, err := dao.GetHelper().GetStatements().GetSQL(repo.StatementCreate)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.Creator", "create statement not found", err)
	}

	return querier.QueryRowContext(ctx, query, params...), nil
}

// changer является внутренним хуком для выполнения SQL-запроса на изменение существующей строки в базе данных.
func (dao *DataAuditOutbox) changer(ctx context.Context, querier db.Querier, entity *extdom.DataAuditOutbox, params ...any) (repo.Scannable, error) {
	query, err := dao.GetHelper().GetStatements().GetSQL(repo.StatementChange)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.Creator", "change statement not found", err)
	}

	return querier.QueryRowContext(ctx, query, params...), nil
}

// pgStatementsBuilder инициализирует и связывает SQL-запросы PostgreSQL с внутренними идентификаторами стейтментов.
func (dao *DataAuditOutbox) pgStatementsBuilder() *repo.Statements[*extdom.DataAuditOutbox, string] {
	return repo.NewStatements[*extdom.DataAuditOutbox, string]().
		WithStatement(repo.StatementFind, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxFind, dao.mapDefault)).
		WithStatement(repo.StatementList, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxList, dao.mapDefault)).
		WithStatement(repo.StatementCreate, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxCreate, dao.mapDefault)).
		WithStatement(repo.StatementChange, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxChange, dao.mapDefault)).
		WithStatement(repo.StatementDelete, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxDelete, dao.mapDefault)).
		WithStatement(statementGetTail, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxGetTail, dao.mapTail)).
		WithStatement(statementListPending, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxListPending, dao.mapDefault)).
		WithStatement(statementListProcessed, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxListProcessed, dao.mapDefault)).
		WithStatement(statementChangeStatus, repo.NewStatement[*extdom.DataAuditOutbox, string](extpg.SqlDataAuditOutboxChangeStatus, nil))
}

// mapDefault преобразует сырые строки базы данных (v2.Scannable) в полноценный объект DataAuditOutbox со всеми заполненными полями.
func (dao *DataAuditOutbox) mapDefault(ctx context.Context, scannable repo.Scannable, params ...any) (*extdom.DataAuditOutbox, error) {
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
func (dao *DataAuditOutbox) mapTail(ctx context.Context, scannable repo.Scannable, params ...any) (*extdom.DataAuditOutbox, error) {
	res := dao.GetHelper().GetHooks().NewEntityFactory()
	err := scannable.Scan(
		&res.ID,
	)
	if err != nil {
		return nil, errs.NewDalError("DataAuditOutbox.mapTail", "tail mapper failed", err)
	}

	return res, nil
}
