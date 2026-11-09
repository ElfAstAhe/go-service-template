package trace

import (
	"context"
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// OutboxListPendingInteractor является декоратором для usecase.OutboxListPendingUseCase,
// предназначенным для автоматического сбора распределенной телеметрии и трассировки.
type OutboxListPendingInteractor struct {
	*telemetry.BaseTelemetry
	spanName string
	next     usecase.OutboxListPendingUseCase
}

var _ usecase.OutboxListPendingUseCase = (*OutboxListPendingInteractor)(nil)

// NewOutboxListPendingUseCase представляет собой фабричный метод для создания
// нового экземпляра трассирующего декоратора OutboxListPendingInteractor.
func NewOutboxListPendingUseCase(
	ucName string,
	next usecase.OutboxListPendingUseCase,
) *OutboxListPendingInteractor {
	return &OutboxListPendingInteractor{
		BaseTelemetry: telemetry.NewBaseTelemetry(ucName),
		spanName:      fmt.Sprintf("%s.List", ucName),
		next:          next,
	}
}

// List выполняет декорируемый вызов для получения списка необработанных элементов domain.Outbox,
// фиксируя при этом параметры лимита, события ошибок и статус текущего спана.
func (olp *OutboxListPendingInteractor) List(ctx context.Context, limit int) ([]domain.Outbox[string], error) {
	ucCtx, span := olp.StartSpan(ctx, olp.spanName)
	defer span.End()

	span.SetAttributes(
		attribute.Int("param.limit", limit),
	)

	res, err := olp.next.List(ucCtx, limit)
	if err != nil {
		span.AddEvent("List_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return res, err
}
