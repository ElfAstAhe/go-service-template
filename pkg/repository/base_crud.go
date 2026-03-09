package repository

import (
	"context"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// BaseCRUDRepository базовая реализация CRUD репозитория
type BaseCRUDRepository[T domain.Entity[ID], ID any] struct {
	queryBuilders *BaseCRUDQueryBuilders
	helper        *Helper[T, ID]
}

func NewBaseCRUDRepository[T domain.Entity[ID], ID any](
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

func (br *BaseCRUDRepository[T, ID]) Find(ctx context.Context, id ID) (T, error) {
	sqlFind, err := br.prepareFind()
	if err != nil {
		return br.GetHelper().GetNilInstance(), err
	}

	return br.GetHelper().Get(ctx, sqlFind, id)
}

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

func (br *BaseCRUDRepository[T, ID]) List(ctx context.Context, limit, offset int) ([]T, error) {
	if err := br.ValidateList(limit, offset); err != nil {
		return nil, errs.NewDalError("BaseCRUDRepository.List", "validate list", err)
	}
	sqlList, err := br.prepareList()
	if err != nil {
		return nil, err
	}

	return br.GetHelper().List(ctx, sqlList, limit, offset)
}

func (br *BaseCRUDRepository[T, ID]) ValidateList(limit, offset int) error {
	if !(limit > 0) {
		return errs.NewDalError("BaseCRUDRepository.ValidateList", "limit must be greater 0", nil)
	}
	if !(offset >= 0) {
		return errs.NewDalError("BaseCRUDRepository.ValidateList", "offset must be equal or greater 0", nil)
	}

	return nil
}

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

func (br *BaseCRUDRepository[T, ID]) Create(ctx context.Context, entity T) (T, error) {
	if err := br.internalValidateCreate(entity); err != nil {
		return br.GetHelper().GetNilInstance(), err
	}

	return br.GetHelper().Create(ctx, entity)
}

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

func (br *BaseCRUDRepository[T, ID]) Change(ctx context.Context, entity T) (T, error) {
	if err := br.internalValidateChange(entity); err != nil {
		return br.GetHelper().GetNilInstance(), err
	}

	return br.GetHelper().Change(ctx, entity)
}

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

func (br *BaseCRUDRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	sqlDelete, err := br.prepareDelete()
	if err != nil {
		return err
	}

	return br.GetHelper().Delete(ctx, sqlDelete, id)
}

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

func (br *BaseCRUDRepository[T, ID]) GetQueryBuilders() *BaseCRUDQueryBuilders {
	return br.queryBuilders
}

func (br *BaseCRUDRepository[T, ID]) GetHelper() *Helper[T, ID] {
	return br.helper
}
