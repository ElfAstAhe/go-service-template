package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// Дополнительные системные маркеры выборки для Owned-репозиториев
const (
	SourceLabelListAll         string = "list_all"
	SourceLabelListAllByOwners string = "list_all_by_owners"
)

// OwnedSaveStrategyManager координирует и выполняет бизнес-логику сохранения зависимых сущностей
// на основе выбранной архитектурной стратегии связывания данных в реляционной СУБД.
type OwnedSaveStrategyManager[T domain.Entity[ID], ID comparable, OwnerID comparable] struct {
	strategies map[LinkStrategy]OwnedSaveFunc[T, ID, OwnerID]
}

// Execute осуществляет ленивый вызов нужной стратегии сохранения, страхуя рантайм от неопределенных конфигураций.
func (ossm *OwnedSaveStrategyManager[T, ID, OwnerID]) Execute(ctx context.Context, strategy LinkStrategy, ownerID OwnerID, owned []T) ([]T, error) {
	if method, ok := ossm.strategies[strategy]; ok {
		return method(ctx, ownerID, owned)
	}

	return nil, errs.NewDalError("OwnedSaveStrategyManager.Execute", fmt.Sprintf("unknown link strategy [%v]", strategy), nil)
}

// BaseOwnedRepository — обобщенная (Generic) базовая реализация репозитория для сущностей,
// жестко связанных со своим владельцем (отношения One-To-Many или Many-To-Many на уровне реляционной схемы).
//
// Полностью удовлетворяет интерфейсу domain.OwnedRepository[T, ID, OwnerID].
type BaseOwnedRepository[T domain.Entity[ID], ID comparable, OwnerID comparable] struct {
	queryBuilders *BaseOwnedQueryBuilders
	helper        *OwnedHelper[T, ID, OwnerID]
	linkStrategy  LinkStrategy
	ownedSaveSM   *OwnedSaveStrategyManager[T, ID, OwnerID]
}

// NewBaseOwnedRepository — фабричный конструктор базового Owned-репозитория.
// Если менеджер стратегий сохранения передан как nil, автоматически собирает пул дефолтных обработчиков.
func NewBaseOwnedRepository[T domain.Entity[ID], ID comparable, OwnerID comparable](
	exec db.Executor,
	errDecipher db.ErrorDecipher,
	info *EntityInfo,
	queryBuilders *BaseOwnedQueryBuilders,
	callbacks *BaseRepositoryCallbacks[T, ID],
	linkStrategy LinkStrategy,
	ownedSaveSM *OwnedSaveStrategyManager[T, ID, OwnerID],
) (*BaseOwnedRepository[T, ID, OwnerID], error) {
	res := &BaseOwnedRepository[T, ID, OwnerID]{
		queryBuilders: queryBuilders,
		helper:        newOwnedHelper[T, ID, OwnerID](exec, errDecipher, callbacks, info),
		linkStrategy:  linkStrategy,
		ownedSaveSM:   ownedSaveSM,
	}
	if ownedSaveSM == nil {
		res.ownedSaveSM = res.buildDefaultSaveStrategies()
	}

	return res, nil
}

// buildDefaultSaveStrategies маппит стандартные методы сохранения OneToMany и ManyToMany во внутреннюю таблицу стратегий.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) buildDefaultSaveStrategies() *OwnedSaveStrategyManager[T, ID, OwnerID] {
	return &OwnedSaveStrategyManager[T, ID, OwnerID]{
		strategies: map[LinkStrategy]OwnedSaveFunc[T, ID, OwnerID]{
			LinkStrategyOneToMany:  bor.saveOneToMany,
			LinkStrategyManyToMany: bor.saveManyToMany,
		},
	}
}

// Find находит и извлекает принадлежащую конкретному владельцу сущность по её составному ключу (ownerID + ID).
func (bor *BaseOwnedRepository[T, ID, OwnerID]) Find(ctx context.Context, ownerID OwnerID, id ID) (T, error) {
	sqlFind, err := bor.prepareFind()
	if err != nil {
		return bor.GetHelper().GetNilInstance(), err
	}

	return bor.GetHelper().Get(ctx, SourceLabelFind, sqlFind, ownerID, id)
}

// prepareFind генерирует и проверяет валидность SQL для поиска по составному ключу.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareFind() (string, error) {
	if bor.queryBuilders == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareFind", "query builders not applied", nil)
	}
	if bor.queryBuilders.GetFind() == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareFind", "query find builder not applied", nil))
	}
	sqlFind := bor.queryBuilders.GetFind()()
	if strings.TrimSpace(sqlFind) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareFind", "query find empty", nil))
	}

	return sqlFind, nil
}

// List возвращает постраничный массив дочерних записей указанного владельца.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) List(ctx context.Context, ownerID OwnerID, limit, offset int) ([]T, error) {
	if err := bor.ValidateList(ownerID, limit, offset); err != nil {
		return nil, err
	}
	sqlList, err := bor.prepareList()
	if err != nil {
		return nil, err
	}

	return bor.GetHelper().List(ctx, SourceLabelList, sqlList, ownerID, limit, offset)
}

// ValidateList верифицирует входящие параметры пагинации.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateList(ownerID OwnerID, limit, offset int) error {
	if !(limit > 0) {
		return errs.NewDalError("BaseOwnedRepository.ValidateList", "limit must be greater 0", nil)
	}
	if !(offset >= 0) {
		return errs.NewDalError("BaseOwnedRepository.ValidateList", "offset must be equal or greater 0", nil)
	}

	return nil
}

// prepareList генерирует SQL для постраничной вычитки owned-коллекций.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareList() (string, error) {
	if bor.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareList", "query builders not applied", nil)
	}
	if bor.GetQueryBuilders().GetList() == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareList", "query find builder not applied", nil))
	}
	sqlList := bor.GetQueryBuilders().GetList()()
	if strings.TrimSpace(sqlList) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareList", "sql list empty", nil))
	}

	return sqlList, nil
}

// ListAll извлекает абсолютно все дочерние записи для одного конкретного владельца.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) ListAll(ctx context.Context, ownerID OwnerID) ([]T, error) {
	if err := bor.ValidateListAll(ownerID); err != nil {
		return nil, err
	}
	sqlList, err := bor.prepareListAll()
	if err != nil {
		return nil, err
	}

	return bor.GetHelper().List(ctx, SourceLabelListAll, sqlList, ownerID)
}

// ValidateListAll верифицирует входящие параметры перед загрузкой списка данных
func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateListAll(ownerID OwnerID) error {
	return nil
}

// prepareListAll формирует SQL-запрос для извлечения полной коллекции записей владельца.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareListAll() (string, error) {
	if bor.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareListAll", "query builders not applied", nil)
	}
	if bor.GetQueryBuilders().GetListAll() == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareListAll", "query find builder not applied", nil))
	}
	sqlListAll := bor.GetQueryBuilders().GetListAll()()
	if strings.TrimSpace(sqlListAll) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareListAll", "sql list empty", nil))
	}

	return sqlListAll, nil
}

// ListAllByOwners осуществляет массовое пакетное извлечение (Bulk Fetching) данных по группе владельцев за один SQL-запрос.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) ListAllByOwners(ctx context.Context, ownerIDs ...OwnerID) (map[OwnerID][]T, error) {
	if err := bor.ValidateListAllByOwners(ownerIDs...); err != nil {
		return nil, err
	}
	sqlListAllByOwners, err := bor.prepareListAllByOwners()
	if err != nil {
		return nil, err
	}

	return bor.GetHelper().ListByOwners(ctx, SourceLabelListAllByOwners, sqlListAllByOwners, ownerIDs)
}

// ValidateListAllByOwners верифицирует входящие пераметры перед загрузкой списка
func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateListAllByOwners(ownerIDs ...OwnerID) error {
	return nil
}

// prepareListAllByOwners формирует SQL для IN-выборки по массиву owner_id.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareListAllByOwners() (string, error) {
	if bor.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareListAllByOwners", "query builders not applied", nil)
	}
	if bor.GetQueryBuilders().GetListAllByOwners() == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareListAllByOwners", "query find builder not applied", nil))
	}
	sqlListAll := bor.GetQueryBuilders().GetListAllByOwners()()
	if strings.TrimSpace(sqlListAll) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareListAllByOwners", "sql list empty", nil))
	}

	return sqlListAll, nil
}

// Save сохраняет, обновляет или синхронизирует состояние owned-коллекции, прозрачно делегируя вызов стейт-менеджеру стратегий.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) Save(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error) {
	return bor.ownedSaveSM.Execute(ctx, bor.linkStrategy, ownerID, owned)
}

// saveManyToMany реализует стратегию полной транзакционной перезаписи (Hard Reset Sync) связей Many-To-Many.
// Метод атомарно сносит старую коллекцию записей владельца и пачкой накатывает новые ассоциации.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) saveManyToMany(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error) {
	// Атомарно удаляем старый пул связей
	if err := bor.DeleteAll(ctx, ownerID); err != nil {
		return nil, err
	}

	// Конвейерно пересоздаем новые ассоциации
	for _, ownedItem := range owned {
		_, err := bor.Create(ctx, ownerID, ownedItem)
		if err != nil {
			return nil, err
		}
	}

	// Возвращаем актуальное состояние коллекции сквозным SELECT-запросом
	return bor.ListAll(ctx, ownerID)
}

// saveOneToMany реализует алгоритм дифференциальной синхронизации коллекции (Diff-Based Synchronization).
// Метод точечно вычисляет дельту на удаление (пропавшие из новой пачки записи), стирает их,
// а оставшиеся марнишрутизирует: UPDATE для IsExists() == true и INSERT для новых объектов.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) saveOneToMany(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error) {
	// Вычисляем массив идентификаторов ID, которые пропали из новой пачки и подлежат удалению
	deleteIDs, err := bor.prepareDeleteList(ctx, ownerID, owned)
	if err != nil {
		return nil, err
	}

	// Точечно вычищаем удаленные сущности
	for _, deleteID := range deleteIDs {
		if err := bor.Delete(ctx, ownerID, deleteID); err != nil {
			return nil, err
		}
	}

	// Синхронизируем состояние оставшейся коллекции
	res := make([]T, 0, len(owned))
	for _, ownedItem := range owned {
		var saved T
		var err error
		if ownedItem.IsExists() {
			// Если запись уже зафиксирована в СУБД — обновляем её поля (UPDATE)
			saved, err = bor.Change(ctx, ownerID, ownedItem)
			if err != nil {
				return nil, errs.NewDalError("BaseOwnedRepository.Save", "error change item", err)
			}
		} else {
			// Если запись абсолютно новая — выполняем вставку (INSERT)
			saved, err = bor.Create(ctx, ownerID, ownedItem)
			if err != nil {
				return nil, errs.NewDalError("BaseOwnedRepository.Save", "error create item", err)
			}
		}
		res = append(res, saved)
	}

	return res, nil
}

// prepareDeleteList за линейное время O(N) вычисляет дельту (разницу) между текущим срезом данных в БД
// и новой входящей пачкой, формируя точный список ID на удаление.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareDeleteList(ctx context.Context, ownerID OwnerID, newItems []T) ([]ID, error) {
	// Индексируем новые ID в хэш-мапу для быстрого O(1) поиска пересечений
	newMap := make(map[ID]struct{}, len(newItems))
	for _, item := range newItems {
		if item.IsExists() {
			newMap[item.GetID()] = struct{}{}
		}
	}

	// Извлекаем текущий слепок данных из СУБД
	existItems, err := bor.ListAll(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	// Если старый ID отсутствует в новой мапе — фиксируем его в дельту на удаление
	res := make([]ID, 0, len(existItems))
	for _, item := range existItems {
		if _, ok := newMap[item.GetID()]; !ok {
			res = append(res, item.GetID())
		}
	}

	return res, nil
}

// Create запускает валидацию и атомарно вставляет новую дочернюю сущность, привязывая её к OwnerID.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) Create(ctx context.Context, ownerID OwnerID, entity T) (T, error) {
	if err := bor.internalValidateCreate(ownerID, entity); err != nil {
		return bor.GetHelper().GetNilInstance(), err
	}

	return bor.GetHelper().Create(ctx, SourceLabelCreate, entity, ownerID)
}

// internalValidateCreate проверяет наличие Creator колбека и верифицирует ValidateCreate на контекст владельца.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) internalValidateCreate(ownerID OwnerID, entity T) error {
	if bor.GetHelper().GetCallbacks().Creator == nil {
		return errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.internalValidateCreate", "creator not applied", nil))
	}

	if bor.GetHelper().GetCallbacks().ValidateCreate != nil {
		if err := bor.GetHelper().GetCallbacks().ValidateCreate(entity, ownerID); err != nil {
			return errs.NewDalError("BaseOwnedRepository.internalValidateCreate", "validate create", err)
		}
	}

	return nil
}

// Change запускает валидацию инвариантов и обновляет поля существующей owned-сущности.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) Change(ctx context.Context, ownerID OwnerID, entity T) (T, error) {
	if err := bor.internalValidateChange(ownerID, entity); err != nil {
		return bor.GetHelper().GetNilInstance(), err
	}

	return bor.GetHelper().Change(ctx, SourceLabelChange, entity)
}

// internalValidateChange проверяет наличие Changer колбека и верифицирует ValidateChange.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) internalValidateChange(ownerID OwnerID, entity T) error {
	if bor.GetHelper().GetCallbacks().Changer == nil {
		return errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.internalValidateChange", "changer not applied", nil))
	}

	if bor.GetHelper().GetCallbacks().ValidateChange != nil {
		if err := bor.GetHelper().GetCallbacks().ValidateChange(entity, ownerID); err != nil {
			return errs.NewDalError("BaseOwnedRepository.internalValidateChange", "validate change", err)
		}
	}

	return nil
}

// DeleteAll каскадно стирает всю коллекцию ресурсов, принадлежащих указанному владельцу ownerID.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) DeleteAll(ctx context.Context, ownerID OwnerID) error {
	if err := bor.ValidateDeleteAll(ownerID); err != nil {
		return errs.NewDalError("BaseOwnedRepository.DeleteAll", "validate delete all", err)
	}
	sqlDeleteAll, err := bor.prepareDeleteAll()
	if err != nil {
		return err
	}

	return bor.GetHelper().DeleteNoCheck(ctx, sqlDeleteAll, ownerID)
}

// ValidateDeleteAll верифицирует входящие параметры перед удалением подчинённых экземпляров
func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateDeleteAll(ownerID OwnerID) error { return nil }

// prepareDeleteAll формирует SQL-запрос для каскадного удаления по owner_id.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareDeleteAll() (string, error) {
	if bor.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareDeleteAll", "query builders not applied", nil)
	}
	if bor.GetQueryBuilders().GetDeleteAll() == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareDeleteAll", "query delete all builder not applied", nil))
	}
	sqlDeleteAll := bor.GetQueryBuilders().GetDeleteAll()()
	if strings.TrimSpace(sqlDeleteAll) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareDeleteAll", "query delete all empty", nil))
	}

	return sqlDeleteAll, nil
}

// Delete удаляет одиночную owned-запись по её ID с жесткой проверкойRowsAffected.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) Delete(ctx context.Context, ownerID OwnerID, id ID) error {
	if err := bor.ValidateDelete(ownerID); err != nil {
		return errs.NewDalError("BaseOwnedRepository.Delete", "validate delete", err)
	}

	sqlDelete, err := bor.prepareDelete()
	if err != nil {
		return err
	}

	return bor.GetHelper().Delete(ctx, sqlDelete, id)
}

// ValidateDelete верифицирует входящие переметры перед удалением экземпляра
func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateDelete(ownerID OwnerID) error { return nil }

// prepareDelete формирует SQL-запрос для удаления записи по первичному ключу.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareDelete() (string, error) {
	if bor.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareDelete", "query builders not applied", nil)
	}
	if bor.GetQueryBuilders().GetDelete() == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareDelete", "query delete builder not applied", nil))
	}
	sqlDelete := bor.GetQueryBuilders().GetDelete()()
	if strings.TrimSpace(sqlDelete) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareDelete", "query delete empty", nil))
	}

	return sqlDelete, nil
}

// GetHelper возвращает ссылку на внутренний инфраструктурный движок выполнения операций.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) GetHelper() *OwnedHelper[T, ID, OwnerID] {
	return bor.helper
}

// GetQueryBuilders возвращает ссылку на зарегистрированный пулл билдеров SQL-запросов.
func (bor *BaseOwnedRepository[T, ID, OwnerID]) GetQueryBuilders() *BaseOwnedQueryBuilders {
	return bor.queryBuilders
}
