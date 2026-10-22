package v2

import (
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

type Hook[E domain.Entity[ID], ID comparable] func(*Hooks[E, ID])

type Hooks[E domain.Entity[ID], ID comparable] struct {
	DispatchMapper   DispatchMapperFunc[E, ID]
	NewEntityFactory NewEntityFactory[E, ID]

	AfterFind      AfterFindFunc[E, ID]
	AfterListYield AfterListYieldFunc[E, ID]

	ValidateCreate ValidateEntityFunc[E, ID]
	BeforeCreate   BeforeCreateFunc[E, ID]
	Creator        CreatorFunc[E, ID]

	ValidateChange ValidateEntityFunc[E, ID]
	BeforeChange   BeforeChangeFunc[E, ID]
	Changer        ChangerFunc[E, ID]
}

func NewHooks[E domain.Entity[ID], ID comparable]() *Hooks[E, ID] {
	return &Hooks[E, ID]{}
}

func (ho *Hooks[E, ID]) Validate() error {
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

func WithCallbackAfterFind[E domain.Entity[ID], ID comparable](callback AfterFindFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.AfterFind = callback
	}
}

func WithCallbackAfterListYield[E domain.Entity[ID], ID comparable](callback AfterListYieldFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.AfterListYield = callback
	}
}

func WithCallbackValidateCreate[E domain.Entity[ID], ID comparable](callback ValidateEntityFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateCreate = callback
	}
}

func WithCallbackBeforeCreate[E domain.Entity[ID], ID comparable](callback BeforeCreateFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.BeforeCreate = callback
	}
}

func WithCallbackCreator[E domain.Entity[ID], ID comparable](callback CreatorFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.Creator = callback
	}
}

func WithCallbackValidateChange[E domain.Entity[ID], ID comparable](callback ValidateEntityFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.ValidateChange = callback
	}
}

func WithCallbackBeforeChange[E domain.Entity[ID], ID comparable](callback BeforeChangeFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.BeforeChange = callback
	}
}

func WithCallbackCHanger[E domain.Entity[ID], ID comparable](callback ChangerFunc[E, ID]) Hook[E, ID] {
	return func(c *Hooks[E, ID]) {
		c.Changer = callback
	}
}
