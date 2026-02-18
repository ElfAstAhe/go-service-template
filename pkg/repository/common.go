package repository

import (
	"context"
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
)

// QueryBuilderFunc билдер sql запроса
type QueryBuilderFunc func() string

// Callback методы базового репозитория
type (
	RowToEntityMapperFunc[T domain.Identity[ID], ID any]  func(row *sql.Row) (*T, error)
	RowsToEntityMapperFunc[T domain.Identity[ID], ID any] func(ctx context.Context, rows *sql.Rows, dest *T) error
	AfterFindFunc[T domain.Identity[ID], ID any]          func(*T) (*T, error)
	AfterListYieldFunc[T domain.Identity[ID], ID any]     func(*T) (*T, error)
	NewEntityFactory[T domain.Identity[ID], ID any]       func() *T
	ValidateEntityFunc[T domain.Identity[ID], ID any]     func(*T) error
	BeforeCreateFunc[T domain.Identity[ID], ID any]       func(*T) error
	BeforeChangeFunc[T domain.Identity[ID], ID any]       func(*T) error
	CreatorFunc[T domain.Identity[ID], ID any]            func(context.Context, *sql.Tx, *T) (*sql.Row, error)
	ChangerFunc[T domain.Identity[ID], ID any]            func(context.Context, *sql.Tx, *T) (*sql.Row, error)
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
