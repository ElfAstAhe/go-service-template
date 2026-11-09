package trace

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/ext/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// OutboxChangeStatusInteractor является декоратором для usecase.OutboxChangeStatusUseCase,
// предназначенным для автоматического сбора распределенной телеметрии и трассировки.
type OutboxChangeStatusInteractor struct {
	*telemetry.BaseTelemetry
	spanName string
	next     usecase.OutboxChangeStatusUseCase
}

var _ usecase.OutboxChangeStatusUseCase = (*OutboxChangeStatusInteractor)(nil)

// NewOutboxChangeStatusUseCase представляет собой фабричный метод для создания
// нового экземпляра трассирующего декоратора OutboxChangeStatusInteractor.
func NewOutboxChangeStatusUseCase(
	ucName string,
	next usecase.OutboxChangeStatusUseCase,
) *OutboxChangeStatusInteractor {
	return &OutboxChangeStatusInteractor{
		BaseTelemetry: telemetry.NewBaseTelemetry(ucName),
		spanName:      fmt.Sprintf("%s.Change", ucName),
		next:          next,
	}
}

// Change выполняет декорируемый вызов для изменения статуса элемента domain.Outbox,
// фиксируя идентификатор записи, целевой статус, события ошибок и статус текущего спана.
func (ocs *OutboxChangeStatusInteractor) Change(ctx context.Context, id string, status string) error {
	ucCtx, span := ocs.StartSpan(ctx, ocs.spanName)
	defer span.End()

	span.SetAttributes(
		attribute.String("param.id", id),
		attribute.String("param.status", status),
	)

	err := ocs.next.Change(ucCtx, id, status)
	if err != nil {
		span.AddEvent("Change_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}
