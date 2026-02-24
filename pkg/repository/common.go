package repository

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// QueryBuilderFunc билдер sql запроса
type QueryBuilderFunc func() string

type Scannable interface {
	Scan(...any) error
}

// Callback методы базового репозитория
type (
	EntityScannerFunc[T domain.Entity[ID], ID any]  func(scanner Scannable, dest T) error
	AfterFindFunc[T domain.Entity[ID], ID any]      func(T) (T, error)
	AfterListYieldFunc[T domain.Entity[ID], ID any] func(T) (T, error)
	NewEntityFactory[T domain.Entity[ID], ID any]   func() T
	ValidateEntityFunc[T domain.Entity[ID], ID any] func(T) error
	BeforeCreateFunc[T domain.Entity[ID], ID any]   func(T) error
	BeforeChangeFunc[T domain.Entity[ID], ID any]   func(T) error
	CreatorFunc[T domain.Entity[ID], ID any]        func(context.Context, *sql.Tx, T) (*sql.Row, error)
	ChangerFunc[T domain.Entity[ID], ID any]        func(context.Context, *sql.Tx, T) (*sql.Row, error)
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
