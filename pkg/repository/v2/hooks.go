package v2

import (
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type Hook[E domain.Entity[ID], ID comparable] func(*Hooks[E, ID])

type Hooks[E domain.Entity[ID], ID comparable] struct {
	DispatchMapper   DispatchMapperFunc[E, ID]
	NewEntityFactory NewEntityFactory[E, ID]

	ValidateFind ValidateIDFunc[ID]
	AfterFind    AfterFindFunc[E, ID]

	ListFilterYield ListFilterYieldFunc[E, ID]

	ValidateCreate ValidateFunc[E, ID]
	BeforeCreate   BeforeFunc[E, ID]
	Creator        ModificatorFunc[E, ID]

	ValidateChange ValidateFunc[E, ID]
	BeforeChange   BeforeFunc[E, ID]
	Changer        ModificatorFunc[E, ID]

	ValidateDelete ValidateIDFunc[ID]
}

func NewHooks[E domain.Entity[ID], ID comparable]() *Hooks[E, ID] {
	return &Hooks[E, ID]{}
}

func (ho *Hooks[E, ID]) Validate() error {
	if utils.IsNil(ho.DispatchMapper) {
		return errs.NewDalError("Hooks[E, ID].Validate", "DispatchMapper required", nil)
	}
	if utils.IsNil(ho.NewEntityFactory) {
		return errs.NewDalError("Hooks[E, ID].Validate", "NewEntityFactory required", nil)
	}

	return nil
}

// ===========================================================
// Fluent API
// ===========================================================

func withHookDispatchMapper[E domain.Entity[ID], ID comparable](dispatcher DispatchMapperFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.DispatchMapper = dispatcher
	}
}

func WithHookNewEntityFactory[E domain.Entity[ID], ID comparable](factory NewEntityFactory[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.NewEntityFactory = factory
	}
}

func WithValidateFind[E domain.Entity[ID], ID comparable](hook ValidateIDFunc[ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateFind = hook
	}
}

func WithHookAfterFind[E domain.Entity[ID], ID comparable](hook AfterFindFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.AfterFind = hook
	}
}

func WithHookListFilterYield[E domain.Entity[ID], ID comparable](hook ListFilterYieldFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ListFilterYield = hook
	}
}

func WithHookValidateCreate[E domain.Entity[ID], ID comparable](hook ValidateFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateCreate = hook
	}
}

func WithHookBeforeCreate[E domain.Entity[ID], ID comparable](hook BeforeFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.BeforeCreate = hook
	}
}

func WithHookCreator[E domain.Entity[ID], ID comparable](hook ModificatorFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.Creator = hook
	}
}

func WithHookValidateChange[E domain.Entity[ID], ID comparable](hook ValidateFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateChange = hook
	}
}

func WithHookBeforeChange[E domain.Entity[ID], ID comparable](hook BeforeFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.BeforeChange = hook
	}
}

func WithHookChanger[E domain.Entity[ID], ID comparable](hook ModificatorFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.Changer = hook
	}
}

func WithHookValidateDelete[E domain.Entity[ID], ID comparable](hook ValidateIDFunc[ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateDelete = hook
	}
}
