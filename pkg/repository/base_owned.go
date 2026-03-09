package repository

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// BaseOwnedRepository one to many or many to many (logic one to many) base implementation repository
type BaseOwnedRepository[T domain.Entity[ID], ID any, OwnerID comparable] struct {
	queryBuilders *BaseOwnedQueryBuilders
	helper        *OwnedHelper[T, ID, OwnerID]
}

func NewBaseOwnedRepository[T domain.Entity[ID], ID any, OwnerID comparable](
	exec db.Executor,
	errDecipher db.ErrorDecipher,
	info *EntityInfo,
	queryBuilders *BaseOwnedQueryBuilders,
	callbacks *BaseRepositoryCallbacks[T, ID],
) (*BaseOwnedRepository[T, ID, OwnerID], error) {
	return &BaseOwnedRepository[T, ID, OwnerID]{
		queryBuilders: queryBuilders,
		helper:        newOwnedHelper[T, ID, OwnerID](exec, errDecipher, callbacks, info),
	}, nil
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) Find(ctx context.Context, ownerID OwnerID, id ID) (T, error) {
	sqlFind, err := bor.prepareFind()
	if err != nil {
		return bor.GetHelper().GetNilInstance(), err
	}

	return bor.GetHelper().Get(ctx, sqlFind, ownerID, id)
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareFind() (string, error) {
	if bor.queryBuilders == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareFind", "query builders not applied", nil)
	}
	if bor.queryBuilders.findBuilder == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareFind", "query find builder not applied", nil))
	}
	sqlFind := bor.queryBuilders.findBuilder()
	if strings.TrimSpace(sqlFind) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareFind", "query find empty", nil))
	}

	return sqlFind, nil
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) List(ctx context.Context, ownerID OwnerID, limit, offset int) ([]T, error) {
	if err := bor.ValidateList(ownerID, limit, offset); err != nil {
		return nil, err
	}
	sqlList, err := bor.prepareList()
	if err != nil {
		return nil, err
	}

	return bor.GetHelper().List(ctx, sqlList, ownerID, limit, offset)
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateList(ownerID OwnerID, limit, offset int) error {
	if !(limit > 0) {
		return errs.NewDalError("BaseOwnedRepository.ValidateList", "limit must be greater 0", nil)
	}
	if !(offset >= 0) {
		return errs.NewDalError("BaseOwnedRepository.ValidateList", "offset must be equal or greater 0", nil)
	}

	return nil
}

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

func (bor *BaseOwnedRepository[T, ID, OwnerID]) ListAll(ctx context.Context, ownerID OwnerID) ([]T, error) {
	if err := bor.ValidateListAll(ownerID); err != nil {
		return nil, err
	}
	sqlList, err := bor.prepareListAll()
	if err != nil {
		return nil, err
	}

	return bor.GetHelper().List(ctx, sqlList, ownerID)
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateListAll(OwnerID OwnerID) error {
	return nil
}

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

func (bor *BaseOwnedRepository[T, ID, OwnerID]) ListAllByOwners(ctx context.Context, ownerIDs ...OwnerID) (map[OwnerID][]T, error) {
	if err := bor.ValidateListAllByOwners(ownerIDs...); err != nil {
		return nil, err
	}
	sqlListAllByOwners, err := bor.prepareListAllByOwners()
	if err != nil {
		return nil, err
	}

	params := make([]any, 0, len(ownerIDs))
	for _, param := range ownerIDs {
		params = append(params, param)
	}

	return bor.GetHelper().ListByOwners(ctx, sqlListAllByOwners, params...)
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateListAllByOwners(ownerIDs ...OwnerID) error {
	return nil
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareListAllByOwners() (string, error) {
	if bor.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareListAllByOwners", "query builders not applied", nil)
	}
	if bor.GetQueryBuilders().GetListAllByOwners() == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareListAllByOwners", "query find builder not applied", nil))
	}
	sqlListAll := bor.GetQueryBuilders().GetListAllByOwners()()
	if strings.TrimSpace(sqlListAll) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareListAllByOnwers", "sql list empty", nil))
	}

	return sqlListAll, nil
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) Save(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error) {
	res := make([]T, 0, len(owned))
	for _, ownedItem := range owned {
		var saved T
		var err error
		if ownedItem.IsExists() {
			saved, err = bor.Change(ctx, ownerID, ownedItem)
			if err != nil {
				return nil, errs.NewDalError("BaseOwnedRepository.Save", "error change item", err)
			}
		} else {
			saved, err = bor.Create(ctx, ownerID, ownedItem)
			if err != nil {
				return nil, errs.NewDalError("BaseOwnedRepository.Save", "error create item", err)
			}
		}
		res = append(res, saved)
	}

	return res, nil
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) Create(ctx context.Context, ownerID OwnerID, entity T) (T, error) {
	if err := bor.internalValidateCreate(ownerID, entity); err != nil {
		return bor.GetHelper().GetNilInstance(), err
	}

	return bor.GetHelper().Create(ctx, entity, ownerID)
}

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

func (bor *BaseOwnedRepository[T, ID, OwnerID]) Change(ctx context.Context, ownerID OwnerID, entity T) (T, error) {
	if err := bor.internalValidateChange(ownerID, entity); err != nil {
		return bor.GetHelper().GetNilInstance(), err
	}

	return bor.GetHelper().Change(ctx, entity)
}

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

func (bor *BaseOwnedRepository[T, ID, OwnerID]) DeleteAll(ctx context.Context, ownerID OwnerID) error {
	if err := bor.ValidateDeleteAll(ownerID); err != nil {
		return errs.NewDalError("BaseOwnedRepository.DeleteAll", "validate delete all", err)
	}
	sqlDeleteAll, err := bor.prepareDeleteAll()
	if err != nil {
		return err
	}

	return bor.GetHelper().Delete(ctx, sqlDeleteAll, ownerID)
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) ValidateDeleteAll(ownerID OwnerID) error {
	return nil
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareDeleteAll() (string, error) {
	if bor.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareDeleteAll", "query builders not applied", nil)
	}
	if bor.GetQueryBuilders().deleteBuilder == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareDeleteAll", "query delete all builder not applied", nil))
	}
	sqlDeleteAll := bor.GetQueryBuilders().deleteBuilder()
	if strings.TrimSpace(sqlDeleteAll) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareDeleteAll", "query delete all empty", nil))
	}

	return sqlDeleteAll, nil
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) Delete(ctx context.Context, ownerID OwnerID, id ID) error {
	sqlDelete, err := bor.prepareDelete()
	if err != nil {
		return err
	}

	return bor.GetHelper().Delete(ctx, sqlDelete, id)
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) prepareDelete() (string, error) {
	if bor.GetQueryBuilders() == nil {
		return "", errs.NewDalError("BaseOwnedRepository.prepareDelete", "query builders not applied", nil)
	}
	if bor.GetQueryBuilders().deleteBuilder == nil {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareDelete", "query delete builder not applied", nil))
	}
	sqlDelete := bor.GetQueryBuilders().deleteBuilder()
	if strings.TrimSpace(sqlDelete) == "" {
		return "", errs.NewNotImplementedError(errs.NewDalError("BaseOwnedRepository.prepareDelete", "query delete empty", nil))
	}

	return sqlDelete, nil
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) GetHelper() *OwnedHelper[T, ID, OwnerID] {
	return bor.helper
}

func (bor *BaseOwnedRepository[T, ID, OwnerID]) GetQueryBuilders() *BaseOwnedQueryBuilders {
	return bor.queryBuilders
}
