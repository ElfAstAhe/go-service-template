package repository

import (
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// BaseRepositoryCallbacks набор методов обратного вызова пост/пред обработки, валидаторов и мэперов
type BaseRepositoryCallbacks[T domain.Entity[ID], ID any] struct {
	RowScanner  RowToEntityMapperFunc[T, ID]
	RowsScanner RowsToEntityMapperFunc[T, ID]

	NewEntityFactory NewEntityFactory[T, ID]
	AfterFind        AfterFindFunc[T, ID]
	AfterListYield   AfterListYieldFunc[T, ID]

	ValidateCreate ValidateEntityFunc[T, ID]
	Creator        CreatorFunc[T, ID]
	BeforeCreate   BeforeCreateFunc[T, ID]

	ValidateChange ValidateEntityFunc[T, ID]
	Changer        ChangerFunc[T, ID]
	BeforeChange   BeforeChangeFunc[T, ID]
}

func newEmptyBaseRepositoryCallbacks[T domain.Entity[ID], ID any]() *BaseRepositoryCallbacks[T, ID] {
	return &BaseRepositoryCallbacks[T, ID]{}
}

type BaseRepositoryCallbacksBuilder[T domain.Entity[ID], ID any] struct {
	instance *BaseRepositoryCallbacks[T, ID]
}

func NewBaseRepositoryCallbacksBuilder[T domain.Entity[ID], ID any]() *BaseRepositoryCallbacksBuilder[T, ID] {
	return &BaseRepositoryCallbacksBuilder[T, ID]{}
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) NewInstance() *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance = newEmptyBaseRepositoryCallbacks[T, ID]()

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithRowMapper(mapper RowToEntityMapperFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.RowScanner = mapper

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithRowsMapper(mapper RowsToEntityMapperFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.RowsScanner = mapper

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithNewEntityFactory(factory NewEntityFactory[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.NewEntityFactory = factory

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithAfterFind(after AfterFindFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.AfterFind = after

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithAfterListYield(after AfterListYieldFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.AfterListYield = after

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithValidateCreate(validate ValidateEntityFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.ValidateCreate = validate

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithCreator(creator CreatorFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.Creator = creator

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithBeforeCreate(before BeforeCreateFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.BeforeCreate = before

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithValidateChange(validate ValidateEntityFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.ValidateChange = validate

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithChanger(changer ChangerFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.Changer = changer

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) WithBeforeChange(before BeforeChangeFunc[T, ID]) *BaseRepositoryCallbacksBuilder[T, ID] {
	bbr.instance.BeforeChange = before

	return bbr
}

func (bbr *BaseRepositoryCallbacksBuilder[T, ID]) Build() (*BaseRepositoryCallbacks[T, ID], error) {
	return bbr.instance, nil
}
