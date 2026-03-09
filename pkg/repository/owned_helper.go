package repository

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type OwnedHelper[T domain.Entity[ID], ID any, OwnerID comparable] struct {
	*Helper[T, ID]
}

func newOwnedHelper[T domain.Entity[ID], ID any, OwnerID comparable](exec db.Executor, errDecipher db.ErrorDecipher, callbacks *BaseRepositoryCallbacks[T, ID], info *EntityInfo) *OwnedHelper[T, ID, OwnerID] {
	return &OwnedHelper[T, ID, OwnerID]{
		Helper: newHelper[T, ID](exec, errDecipher, callbacks, info),
	}
}

func (oh *OwnedHelper[T, ID, OwnerID]) ListByOwners(ctx context.Context, sqlReq string, params ...any) (map[OwnerID][]T, error) {
	querier := oh.GetExecutor().GetQuerier(ctx)

	rows, err := querier.QueryContext(ctx, sqlReq, params...)
	if err != nil {
		return nil, errs.NewDalError("OwnedHelper.List", "query", err)
	}
	defer rows.Close()

	res := make(map[OwnerID][]T)
	for rows.Next() {
		if err = ctx.Err(); err != nil {
			return nil, errs.NewDalError("OwnedHelper.List", "check context", err)
		}

		addEntity := true
		var ownerID OwnerID
		entity := oh.GetCallbacks().NewEntityFactory()

		err = oh.GetCallbacks().EntityScanner(rows, entity, ownerID)
		if err != nil {
			return nil, errs.NewDalError("OwnedHelper.ListByOwners", "scan rows", err)
		}

		if oh.GetCallbacks().AfterListYield != nil {
			entity, addEntity, err = oh.GetCallbacks().AfterListYield(entity, ownerID)
			if err != nil {
				return nil, errs.NewDalError("OwnedHelper.List", "post scan processing", err)
			}
		}
		// yeld метод постобработки строки не вернул entity, нет данных - нет добавления
		if any(entity) == nil || !addEntity {
			continue
		}
		if _, ok := res[ownerID]; !ok {
			res[ownerID] = make([]T, 0)
		}
		res[ownerID] = append(res[ownerID], entity)
	}
	if rows.Err() != nil {
		return nil, errs.NewDalError("OwnedHelper.List", "after scan", rows.Err())
	}

	return res, nil
}
