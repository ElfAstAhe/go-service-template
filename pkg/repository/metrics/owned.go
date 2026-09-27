package metrics

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// BaseOwnedMetricsRepository — универсальный дженерик-декоратор (обертка) для репозиториев с разделением прав (OwnedRepository).
//
// Полностью реализует интерфейс domain.OwnedRepository[T, ID, OwnerID].
// Обеспечивает сквозное профилирование транзакций и SQL-запросов, завязанных на сущность-владельца (Owner),
// сохраняя при этом строгое разделение ответственности (SRP) и чистоту кода самого слоя доступа к данным.
type BaseOwnedMetricsRepository[T domain.Entity[ID], ID comparable, OwnerID comparable] struct {
	repository domain.OwnedRepository[T, ID, OwnerID] // Ссылка на реальный нижележащий репозиторий баз данных
	repoName   string                                 // Уникальное имя репозитория для label-тега в Prometheus
}

// NewBaseOwnedMetricsRepository — фабричный конструктор декоратора Owned-репозитория.
// Автоматически вычисляет строковое имя типа через рефлексию (utils.GetTypeName), если repoName передан пустым.
func NewBaseOwnedMetricsRepository[T domain.Entity[ID], ID comparable, OwnerID comparable](repoName string, repository domain.OwnedRepository[T, ID, OwnerID]) *BaseOwnedMetricsRepository[T, ID, OwnerID] {
	res := &BaseOwnedMetricsRepository[T, ID, OwnerID]{
		repoName:   repoName,
		repository: repository,
	}
	// Проверка на пустоту (== "") гарантирует сохранение кастомного repoName
	if repoName == "" {
		res.repoName = utils.GetTypeName(repository)
	}

	return res
}

// Find — декорированный метод поиска принадлежащей владельцу сущности по ID.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Find(ctx context.Context, ownerID OwnerID, id ID) (res T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "Find", err, start)
	}(time.Now())

	return omr.repository.Find(ctx, ownerID, id)
}

// List — декорированный метод постраничной выборки сущностей конкретного владельца.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) List(ctx context.Context, ownerID OwnerID, limit, offset int) (res []T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "List", err, start)
	}(time.Now())

	return omr.repository.List(ctx, ownerID, limit, offset)
}

// ListAll — декорированный метод полной выборки всех сущностей владельца.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) ListAll(ctx context.Context, ownerID OwnerID) (res []T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "ListAll", err, start)
	}(time.Now())

	return omr.repository.ListAll(ctx, ownerID)
}

// ListAllByOwners — декорированный метод пакетного извлечения данных по группе владельцев (Bulk Select).
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) ListAllByOwners(ctx context.Context, ownerIDs ...OwnerID) (res map[OwnerID][]T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "ListAllByOwners", err, start)
	}(time.Now())

	return omr.repository.ListAllByOwners(ctx, ownerIDs...)
}

// Save — декорированный метод пакетного сохранения/обновления связанных сущностей.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Save(ctx context.Context, ownerID OwnerID, owned []T) (res []T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "Save", err, start)
	}(time.Now())

	return omr.repository.Save(ctx, ownerID, owned)
}

// Create — декорированный метод создания новой дочерней сущности.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Create(ctx context.Context, ownerID OwnerID, entity T) (res T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "Create", err, start)
	}(time.Now())

	return omr.repository.Create(ctx, ownerID, entity)
}

// Change — декорированный метод изменения параметров принадлежащей владельцу сущности.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Change(ctx context.Context, ownerID OwnerID, entity T) (res T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "Change", err, start)
	}(time.Now())

	return omr.repository.Change(ctx, ownerID, entity)
}

// DeleteAll — декорированный метод полного каскадного удаления всех ресурсов владельца.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) DeleteAll(ctx context.Context, ownerID OwnerID) (err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "DeleteAll", err, start)
	}(time.Now())

	return omr.repository.DeleteAll(ctx, ownerID)
}

// Delete — декорированный метод атомарного удаления конкретной сущности владельца.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) Delete(ctx context.Context, ownerID OwnerID, id ID) (err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(omr.repoName, "Delete", err, start)
	}(time.Now())

	return omr.repository.Delete(ctx, ownerID, id)
}

// GetRepositoryName возвращает зафиксированное имя репозитория.
func (omr *BaseOwnedMetricsRepository[T, ID, OwnerID]) GetRepositoryName() string {
	return omr.repoName
}
