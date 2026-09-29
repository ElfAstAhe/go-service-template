package repository

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// Системные константы-маркеры (Labels) для идентификации вызывающих методов
// внутри логов, метрик и трассировщиков инфраструктурного хелпера.
const (
	SourceLabelFind   string = "find"
	SourceLabelList   string = "list"
	SourceLabelCreate string = "create"
	SourceLabelChange string = "change"
)

// BaseCRUDRepository — обобщенная (Generic) реализация базового CRUD-репозитория.
//
// Удовлетворяет интерфейсу domain.CRUDRepository[T, ID].
// Пакует в себя две ключевые абстракции:
// 1. BaseCRUDQueryBuilders: отвечает исключительно за декларативную генерацию чистых SQL-запросов.
// 2. Helper[T, ID]: оборачивает рантайм выполнения SQL, управление Querier/Транзакциями и маппинг строк.
type BaseCRUDRepository[T domain.Entity[ID], ID comparable] struct {
	queryBuilders *BaseCRUDQueryBuilders
	helper        *Helper[T, ID]
}

// NewBaseCRUDRepository — фабричный конструктор базового CRUD-репозитория.
// Инициализирует внутренний мост Helper, связывая executor, дешифратор ошибок и жизненные циклы колбеков.
func NewBaseCRUDRepository[T domain.Entity[ID], ID comparable](
	exec db.Executor,
	errDecipher db.ErrorDecipher,
	info *EntityInfo,
	queryBuilders *BaseCRUDQueryBuilders,
	callbacks *BaseRepositoryCallbacks[T, ID],
) (*BaseCRUDRepository[T, ID], error) {
	return &BaseCRUDRepository[T, ID]{
		queryBuilders: queryBuilders,
		helper:        newHelper[T, ID](exec, errDecipher, callbacks, info),
	}, nil
}

// Find выполняет поиск и извлечение одиночной доменной сущности по её уникальному идентификатору.
func (br *BaseCRUDRepository[T, ID]) Find(ctx context.Context, id ID) (T, error) {
	sqlFind, err := br.prepareFind()
	if err != nil {
		return br.GetHelper().GetNilInstance(), err
	}

	return br.GetHelper().Get(ctx, SourceLabelFind, sqlFind, id)
}

// prepareFind генерирует и валидирует SQL-строку для операции чтения (SELECT BY ID).
func (br *BaseCRUDRepository[T, ID]) prepareFind() (string, error) {
	if br.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseCRUDRepository.prepareFind", "query builders not applied", nil)
	}
	if br.GetQueryBuilders().findBuilder == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseCRUDRepository.prepareFind", "query find builder not applied", nil))
	}
	sqlFind := br.GetQueryBuilders().findBuilder()
	if strings.TrimSpace(sqlFind) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseCRUDRepository.prepareFind", "query find empty", nil))
	}

	return sqlFind, nil
}

// List возвращает постраничную выборку (срез) доменных сущностей с учетом лимитов и смещений.
func (br *BaseCRUDRepository[T, ID]) List(ctx context.Context, limit, offset int) ([]T, error) {
	if err := br.ValidateList(limit, offset); err != nil {
		return nil, errs.NewDalError("BaseCRUDRepository.List", "validate list", err)
	}
	sqlList, err := br.prepareList()
	if err != nil {
		return nil, err
	}

	return br.GetHelper().List(ctx, SourceLabelList, sqlList, limit, offset)
}

// ValidateList проверяет корректность параметров пагинации перед отправкой запроса в СУБД.
func (br *BaseCRUDRepository[T, ID]) ValidateList(limit, offset int) error {
	if limit <= 0 {
		return errs.NewDalError("BaseCRUDRepository.ValidateList", "limit must be greater 0", nil)
	}
	if offset <= 0 {
		return errs.NewDalError("BaseCRUDRepository.ValidateList", "offset must be equal or greater 0", nil)
	}

	return nil
}

// prepareList генерирует и валидирует SQL-строку для выборки массивов (SELECT LIST).
func (br *BaseCRUDRepository[T, ID]) prepareList() (string, error) {
	if br.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseCRUDRepository.prepareList", "query builders not applied", nil)
	}
	if br.GetQueryBuilders().GetList() == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseCRUDRepository.prepareList", "query find builder not applied", nil))
	}
	sqlList := br.GetQueryBuilders().GetList()()
	if strings.TrimSpace(sqlList) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseCRUDRepository.prepareList", "sql list empty", nil))
	}

	return sqlList, nil
}

// Create запускает валидацию бизнес-инвариантов и атомарно вставляет новую сущность в базу данных.
func (br *BaseCRUDRepository[T, ID]) Create(ctx context.Context, entity T) (T, error) {
	if err := br.internalValidateCreate(entity); err != nil {
		return br.GetHelper().GetNilInstance(), err
	}

	return br.GetHelper().Create(ctx, SourceLabelCreate, entity)
}

// internalValidateCreate страхует рантайм от паник, проверяя наличие Creator колбека и вызывая ValidateCreate.
func (br *BaseCRUDRepository[T, ID]) internalValidateCreate(entity T) error {
	if br.GetHelper().GetCallbacks().Creator == nil {
		return errs.NewNotImplementedError(errs.NewDalError("BaseCRUDRepository.internalValidateCreate", "creator not applied", nil))
	}

	if br.GetHelper().GetCallbacks().ValidateCreate != nil {
		if err := br.GetHelper().GetCallbacks().ValidateCreate(entity); err != nil {
			return errs.NewDalError("BaseCRUDRepository.internalValidateCreate", "validate create", err)
		}
	}

	return nil
}

// Change выполняет обновление (UPDATE) полей существующей сущности.
func (br *BaseCRUDRepository[T, ID]) Change(ctx context.Context, entity T) (T, error) {
	if err := br.internalValidateChange(entity); err != nil {
		return br.GetHelper().GetNilInstance(), err
	}

	return br.GetHelper().Change(ctx, SourceLabelChange, entity)
}

// internalValidateChange страхует рантайм от паник, проверяя наличие Changer колбека и вызывая ValidateChange.
func (br *BaseCRUDRepository[T, ID]) internalValidateChange(entity T) error {
	if br.GetHelper().GetCallbacks().Changer == nil {
		return errs.NewNotImplementedError(errs.NewDalError("BaseCRUDRepository.internalValidateChange", "changer not applied", nil))
	}

	if br.GetHelper().GetCallbacks().ValidateChange != nil {
		if err := br.GetHelper().GetCallbacks().ValidateChange(entity); err != nil {
			return errs.NewDalError("BaseCRUDRepository.internalValidateChange", "validate change", err)
		}
	}

	return nil
}

// Delete удаляет запись из физической таблицы по её уникальному идентификатору ID.
func (br *BaseCRUDRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	sqlDelete, err := br.prepareDelete()
	if err != nil {
		return err
	}

	return br.GetHelper().Delete(ctx, sqlDelete, id)
}

// prepareDelete генерирует и валидирует SQL-строку для операции удаления (DELETE FROM).
func (br *BaseCRUDRepository[T, ID]) prepareDelete() (string, error) {
	if br.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseCRUDRepository.prepareDelete", "query builders not applied", nil)
	}
	if br.GetQueryBuilders().deleteBuilder == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseCRUDRepository.prepareDelete", "query delete builder not applied", nil))
	}
	sqlDelete := br.GetQueryBuilders().deleteBuilder()
	if strings.TrimSpace(sqlDelete) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseCRUDRepository.prepareDelete", "query delete empty", nil))
	}

	return sqlDelete, nil
}

// GetQueryBuilders возвращает ссылку на зарегистрированный пулл билдеров SQL-запросов.
func (br *BaseCRUDRepository[T, ID]) GetQueryBuilders() *BaseCRUDQueryBuilders {
	return br.queryBuilders
}

// GetHelper возвращает ссылку на внутренний инфраструктурный движок выполнения операций.
func (br *BaseCRUDRepository[T, ID]) GetHelper() *Helper[T, ID] {
	return br.helper
}
