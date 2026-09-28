package trace

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/telemetry"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// BaseCRUDTraceRepository — универсальный дженерик-декоратор для распределенной трассировки CRUD-операций.
//
// Реализует контракт domain.CRUDRepository[T, ID].
// Инкапсулирует работу с OpenTelemetry (OTel), прозрачно порождая дочерние спаны (Spans)
// для визуализации графа вызовов, записи ошибок баз данных и фиксации метаданных выполнения (Jaeger/OpenSearch).
type BaseCRUDTraceRepository[T domain.Entity[ID], ID comparable] struct {
	*telemetry.BaseTelemetry                              // Встраиваем базовые механизмы трассировки и логгирования фреймворка
	repository               domain.CRUDRepository[T, ID] // Ссылка на нижележащий (или следующий в цепочке) репозиторий
	nilEntity                T                            // Нулевой литерал сущности для безопасного возврата при ошибках
}

// NewBaseCRUDTraceRepository — фабричный конструктор декоратора трассировки.
// Если repositoryName пуст, имя трасера автоматически вычисляется через имя типа нативной структуры (utils.GetTypeName).
func NewBaseCRUDTraceRepository[T domain.Entity[ID], ID comparable](repositoryName string, repository domain.CRUDRepository[T, ID]) *BaseCRUDTraceRepository[T, ID] {
	res := &BaseCRUDTraceRepository[T, ID]{
		repository: repository,
	}
	tracerName := repositoryName
	if tracerName == "" {
		tracerName = utils.GetTypeName(repository)
	}
	res.BaseTelemetry = telemetry.NewBaseTelemetry(tracerName)

	return res
}

// Find — декорированный метод поиска записи с открытием именованного спана и фиксацией искомого ID.
func (btr *BaseCRUDTraceRepository[T, ID]) Find(ctx context.Context, id ID) (T, error) {
	ctx, span := btr.StartSpan(ctx, fmt.Sprintf("%s.Find", btr.GetRepositoryName()))
	defer span.End()

	// Обогащаем спан семантическими атрибутами для удобной фильтрации в UI визуализатора
	span.SetAttributes(attribute.String("param.id", fmt.Sprintf("%v", id)))

	res, err := btr.repository.Find(ctx, id)
	if err != nil {
		// Фиксируем событие сбоя и маршалируем трейс ошибки в OTel контекст
		span.AddEvent("Find_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return btr.GetNilEntity(), err
	}

	return res, nil
}

// List — декорированный метод трассировки вывода постраничных списков с записью лимитов и смещений.
func (btr *BaseCRUDTraceRepository[T, ID]) List(ctx context.Context, limit, offset int) ([]T, error) {
	ctx, span := btr.GetTracer().Start(ctx, fmt.Sprintf("%s.List", btr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(
		attribute.Int("param.limit", limit),
		attribute.Int("param.offset", offset),
	)

	res, err := btr.repository.List(ctx, limit, offset)
	if err != nil {
		span.AddEvent("List_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return res, nil
}

// Create — декорированный метод создания записи. Логирует сгенерированный или переданный ID сущности.
func (btr *BaseCRUDTraceRepository[T, ID]) Create(ctx context.Context, entity T) (T, error) {
	ctx, span := btr.StartSpan(ctx, fmt.Sprintf("%s.Create", btr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(attribute.String("param.entity_id", fmt.Sprintf("%v", entity.GetID())))

	res, err := btr.repository.Create(ctx, entity)
	if err != nil {
		span.AddEvent("Create_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return btr.GetNilEntity(), err
	}

	return res, nil
}

// Change — декорированный метод модификации записи.
func (btr *BaseCRUDTraceRepository[T, ID]) Change(ctx context.Context, entity T) (T, error) {
	ctx, span := btr.StartSpan(ctx, fmt.Sprintf("%s.Change", btr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(attribute.String("param.entity_id", fmt.Sprintf("%v", entity.GetID())))

	res, err := btr.repository.Change(ctx, entity)
	if err != nil {
		span.AddEvent("Change_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return btr.GetNilEntity(), err
	}

	return res, nil
}

// Delete — декорированный метод удаления записи.
func (btr *BaseCRUDTraceRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	ctx, span := btr.StartSpan(ctx, fmt.Sprintf("%s.Delete", btr.GetRepositoryName()))
	defer span.End()

	span.SetAttributes(attribute.String("param.id", fmt.Sprintf("%v", id)))

	err := btr.repository.Delete(ctx, id)
	if err != nil {
		span.AddEvent("Delete_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

// GetRepositoryName возвращает строковый идентификатор трассируемого репозитория.
func (btr *BaseCRUDTraceRepository[T, ID]) GetRepositoryName() string {
	return btr.GetTracerName()
}

// GetNilEntity возвращает пустой литерал структуры T для корректного и безопасного выхода.
func (btr *BaseCRUDTraceRepository[T, ID]) GetNilEntity() T {
	return btr.nilEntity
}
