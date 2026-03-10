package repository

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

type LinkStrategy int

const (
	LinkStrategyOneToMany LinkStrategy = iota
	LinkStrategyManyToMany
)

type OwnedSaveFunc[T domain.Entity[ID], ID comparable, OwnerID comparable] func(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error)

// QueryBuilderFunc билдер sql запроса
type QueryBuilderFunc func() string

type Scannable interface {
	Scan(...any) error
}

// Callback методы базового репозитория
type (
	EntityScannerFunc[T domain.Entity[ID], ID comparable]  func(scanner Scannable, dest T, params ...any) error
	AfterFindFunc[T domain.Entity[ID], ID comparable]      func(T, ...any) (T, error)
	AfterListYieldFunc[T domain.Entity[ID], ID comparable] func(T, ...any) (T, bool, error)
	NewEntityFactory[T domain.Entity[ID], ID comparable]   func() T
	ValidateEntityFunc[T domain.Entity[ID], ID comparable] func(T, ...any) error
	BeforeCreateFunc[T domain.Entity[ID], ID comparable]   func(T, ...any) error
	BeforeChangeFunc[T domain.Entity[ID], ID comparable]   func(T, ...any) error
	CreatorFunc[T domain.Entity[ID], ID comparable]        func(context.Context, db.Querier, T, ...any) (*sql.Row, error)
	ChangerFunc[T domain.Entity[ID], ID comparable]        func(context.Context, db.Querier, T, ...any) (*sql.Row, error)
)

type EntityInfo struct {
	Table  string
	Entity string
}

func NewEntityInfo(table, entity string) *EntityInfo {
	return &EntityInfo{
		Table:  table,
		Entity: entity,
	}
}
