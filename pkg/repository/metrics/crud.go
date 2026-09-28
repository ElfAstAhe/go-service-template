package metrics

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// BaseCRUDMetricsRepository — универсальный дженерик-декоратор (обертка) для любого CRUD-репозитория фреймворка.
//
// Полностью реализует интерфейс domain.CRUDRepository[T, ID].
// Обеспечивает строгое соблюдение принципа Single Responsibility (SRP):
// Нативный репозиторий (например, в PgContainer) отвечает исключительно за сырые SQL-запросы,
// а данный декоратор прозрачно оборачивает его методы для автоматического сбора телеметрии Prometheus.
type BaseCRUDMetricsRepository[T domain.Entity[ID], ID comparable] struct {
	repository domain.CRUDRepository[T, ID] // Ссылка на реальный нижележащий репозиторий баз данных
	repoName   string                       // Уникальное имя репозитория для label-тега в Prometheus
}

// NewBaseCRUDMetricsRepository — фабричный конструктор декоратора репозитория.
// Автоматически вычисляет строковое имя типа через рефлексию (utils.GetTypeName), если repoName передан пустым.
func NewBaseCRUDMetricsRepository[T domain.Entity[ID], ID comparable](repoName string, repository domain.CRUDRepository[T, ID]) *BaseCRUDMetricsRepository[T, ID] {
	res := &BaseCRUDMetricsRepository[T, ID]{
		repository: repository,
		repoName:   repoName,
	}
	if repoName == "" {
		res.repoName = utils.GetTypeName(repository)
	}

	return res
}

// Find — декорированный метод поиска сущности по ID с автоматическим трекингом Latency и ошибок.
func (bmr *BaseCRUDMetricsRepository[T, ID]) Find(ctx context.Context, id ID) (res T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(bmr.repoName, "Find", err, start)
	}(time.Now())

	return bmr.repository.Find(ctx, id)
}

// List — декорированный метод постраничной выборки сущностей.
func (bmr *BaseCRUDMetricsRepository[T, ID]) List(ctx context.Context, limit, offset int) (res []T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(bmr.repoName, "List", err, start)
	}(time.Now())

	return bmr.repository.List(ctx, limit, offset)
}

// Create — декорированный метод вставки новой записи в базу данных.
func (bmr *BaseCRUDMetricsRepository[T, ID]) Create(ctx context.Context, entity T) (res T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(bmr.repoName, "Create", err, start)
	}(time.Now())

	return bmr.repository.Create(ctx, entity)
}

// Change — декорированный метод обновления существующей сущности.
func (bmr *BaseCRUDMetricsRepository[T, ID]) Change(ctx context.Context, entity T) (res T, err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(bmr.repoName, "Change", err, start)
	}(time.Now())

	return bmr.repository.Change(ctx, entity)
}

// Delete — декорированный метод удаления записи по её идентификатору.
func (bmr *BaseCRUDMetricsRepository[T, ID]) Delete(ctx context.Context, id ID) (err error) {
	defer func(start time.Time) {
		ObserveRepositoryOp(bmr.repoName, "Delete", err, start)
	}(time.Now())

	return bmr.repository.Delete(ctx, id)
}

// GetRepositoryName возвращает зафиксированное имя репозитория.
func (bmr *BaseCRUDMetricsRepository[T, ID]) GetRepositoryName() string {
	return bmr.repoName
}
