package domain

import (
	"context"
)

type Repository[T Entity[ID], ID any] interface {
	Find(ctx context.Context, id ID) (T, error)

	List(ctx context.Context, limit, offset int) ([]T, error)

	Create(ctx context.Context, entity T) (T, error)
	Change(ctx context.Context, entity T) (T, error)

	Delete(ctx context.Context, id ID) error

	Close() error
}
