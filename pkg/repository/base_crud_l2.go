package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type BaseCRUDL2Repository[E domain.Entity[ID], ID comparable] struct {
	next       domain.CRUDRepository[E, ID]
	entityInfo *EntityInfo
	crudCache  cache.Cache[ID, E]
	nilEntity  E
	defaultTTL time.Duration
	log        logger.Logger
}

func NewBaseCRUDL2Repository[E domain.Entity[ID], ID comparable](
	next domain.CRUDRepository[E, ID],
	entityInfo *EntityInfo,
	crudCache cache.Cache[ID, E],
	defaultTTL time.Duration,
	log logger.Logger,
) *BaseCRUDL2Repository[E, ID] {
	return &BaseCRUDL2Repository[E, ID]{
		next:       next,
		entityInfo: entityInfo,
		crudCache:  crudCache,
		defaultTTL: defaultTTL,
		log:        log.GetLogger("BaseCRUDL2Repository"),
	}
}

func (bcl *BaseCRUDL2Repository[E, ID]) Find(ctx context.Context, id ID) (E, error) {
	// from cache
	res, ok, err := bcl.crudCache.Get(id)
	if err != nil {
		return bcl.nilEntity, errs.NewDalCacheError("BaseCRUDL2Repository.Find", fmt.Sprintf("get from cache entity id [%v]", id), err)
	}
	if ok {
		if utils.IsNil(res) {
			return res, errs.NewDalNotFoundError(bcl.GetInfo().Entity, "not found", nil)
		}

		return res, nil
	}
	// orig op
	res, err = bcl.next.Find(ctx, id)
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); !ok {
			return res, err
		}
	}
	// put into cache
	cacheErr := bcl.crudCache.Set(id, res, bcl.defaultTTL)
	if cacheErr != nil {
		bcl.log.Errorf(fmt.Sprintf("set into cache entity id [%v]", id), cacheErr)
	}

	return res, err
}

func (bcl *BaseCRUDL2Repository[E, ID]) List(ctx context.Context, limit, offset int) ([]E, error) {
	// orig op
	return bcl.next.List(ctx, limit, offset)
}

func (bcl *BaseCRUDL2Repository[E, ID]) Create(ctx context.Context, entity E) (E, error) {
	// orig op
	res, err := bcl.next.Create(ctx, entity)
	if err != nil {
		return bcl.nilEntity, err
	}
	// put into cache
	cacheErr := bcl.crudCache.Set(res.GetID(), res, bcl.defaultTTL)
	if cacheErr != nil {
		bcl.log.Errorf(fmt.Sprintf("set into cache entity id [%v]", res.GetID()), cacheErr)
	}

	return res, err
}

func (bcl *BaseCRUDL2Repository[E, ID]) Change(ctx context.Context, entity E) (E, error) {
	// orig op
	res, err := bcl.next.Change(ctx, entity)
	if err != nil {
		if _, ok := errors.AsType[*errs.DalNotFoundError](err); !ok {
			return res, err
		}
	}
	// put into cache
	cacheErr := bcl.crudCache.Set(res.GetID(), res, bcl.defaultTTL)
	if cacheErr != nil {
		bcl.log.Errorf(fmt.Sprintf("set into cache entity id [%v]", res.GetID()), cacheErr)
	}

	return res, err
}

func (bcl *BaseCRUDL2Repository[E, ID]) Delete(ctx context.Context, id ID) error {
	err := bcl.next.Delete(ctx, id)
	if err != nil {
		return err
	}
	// delete from crud cache
	bcl.crudCache.Delete(id)

	return nil
}

func (bcl *BaseCRUDL2Repository[E, ID]) GetInfo() *EntityInfo {
	return bcl.entityInfo
}

func (bcl *BaseCRUDL2Repository[E, ID]) GetCache() cache.Cache[ID, E] {
	return bcl.crudCache
}

func (bcl *BaseCRUDL2Repository[E, ID]) GetDefaultTTL() time.Duration {
	return bcl.defaultTTL
}

func (bcl *BaseCRUDL2Repository[E, ID]) GetLogger() logger.Logger {
	return bcl.log
}
