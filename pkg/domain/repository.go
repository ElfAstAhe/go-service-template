package domain

import (
	"context"
)

// CrudRepository repository with simple crud methods for nonowned instances
type CrudRepository[T Entity[ID], ID any] interface {
	Find(ctx context.Context, id ID) (T, error)

	List(ctx context.Context, limit, offset int) ([]T, error)

	Create(ctx context.Context, entity T) (T, error)
	Change(ctx context.Context, entity T) (T, error)

	Delete(ctx context.Context, id ID) error
}

// OwnedRepository repository with simple crud methods for owned instances
type OwnedRepository[T Entity[ID], ID any, OwnerID comparable] interface {
	Find(ctx context.Context, ownerID OwnerID, id ID) (T, error)

	List(ctx context.Context, ownerID OwnerID, limit, offset int) ([]T, error)
	ListAll(ctx context.Context, ownerID OwnerID) ([]T, error)
	ListAllByOwners(ctx context.Context, ownerIDs ...OwnerID) (map[OwnerID][]T, error)

	Save(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error)
	Create(ctx context.Context, ownerID OwnerID, entity T) (T, error)
	Change(ctx context.Context, ownerID OwnerID, entity T) (T, error)

	DeleteAll(ctx context.Context, ownerID OwnerID) error
	Delete(ctx context.Context, ownerID OwnerID, id ID) error
}
