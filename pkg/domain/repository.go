package domain

import (
	"context"
)

// CRUDRepository описывает универсальный, строго типизированный (Generic) контракт
// базового репозитория для независимых доменных сущностей (Плоский CRUD).
//
// Инкапсулирует каноничные атомарные операции работы с персистентным хранилищем СУБД.
type CRUDRepository[T Entity[ID], ID comparable] interface {
	// Find осуществляет точечный поиск и извлечение одиночной сущности по её первичному ключу ID.
	Find(ctx context.Context, id ID) (T, error)

	// List возвращает постраничный массив (срез) сущностей с учетом лимитов и смещений пагинации.
	List(ctx context.Context, limit, offset int) ([]T, error)

	// Create выполняет валидацию и атомарную фиксацию (INSERT) новой сущности в СУБД.
	Create(ctx context.Context, entity T) (T, error)

	// Change производит обновление (UPDATE) полей существующей в памяти сущности.
	Change(ctx context.Context, entity T) (T, error)

	// Delete принудительно удаляет запись из хранилища по её уникальному идентификатору ID.
	Delete(ctx context.Context, id ID) error
}

// OwnedRepository описывает универсальный, строго типизированный (Generic) контракт репозитория
// для дочерних (зависимых) сущностей, жестко привязанных к контексту сущности-владельца (OwnerID).
//
// Поддерживает каскадные операции синхронизации пачек и пакетного извлечения данных (Bulk Select).
type OwnedRepository[T Entity[ID], ID comparable, OwnerID comparable] interface {
	// Find находит и извлекает принадлежащую конкретному владельцу сущность по составному ключу (ownerID + id).
	Find(ctx context.Context, ownerID OwnerID, id ID) (T, error)

	// List возвращает постраничную коллекцию дочерних записей указанного владельца.
	List(ctx context.Context, ownerID OwnerID, limit, offset int) ([]T, error)

	// ListAll извлекает абсолютно все дочерние записи, закрепленные за конкретным владельцем ownerID.
	ListAll(ctx context.Context, ownerID OwnerID) ([]T, error)

	// ListAllByOwners осуществляет пакетное извлечение данных (Bulk Fetching) по группе владельцев за один SQL-запрос.
	// Оптимизирует ввод-вывод СУБД, превентивно блокируя возникновение архитектурной проблемы N+1.
	ListAllByOwners(ctx context.Context, ownerIDs ...OwnerID) (map[OwnerID][]T, error)

	// Save координирует запуск алгоритмов дифференциальной синхронизации или полной перезаписи
	// зависимых коллекций (One-To-Many / Many-To-Many) в рамках единой ACID-транзакции.
	Save(ctx context.Context, ownerID OwnerID, owned []T) ([]T, error)

	// Create создает и привязывает новую дочернюю сущность к указанному OwnerID.
	Create(ctx context.Context, ownerID OwnerID, entity T) (T, error)

	// Change обновляет параметры существующей дочерней сущности в границах контекста владельца.
	Change(ctx context.Context, ownerID OwnerID, entity T) (T, error)

	// DeleteAll каскадно вычищает всю коллекцию дочерних ресурсов, принадлежащих указанному владельцу.
	DeleteAll(ctx context.Context, ownerID OwnerID) error

	// Delete производит точечное удаление связанного ресурса по его ID и OwnerID.
	Delete(ctx context.Context, ownerID OwnerID, id ID) error
}
