package trace

import (
	"context"
	"fmt"

	extuc "github.com/ElfAstAhe/go-service-template/pkg/ext/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// TailCutInteractor является универсальным (generic) декоратором для usecase.TailCutUseCase,
// предназначенным для автоматического сбора распределенной телеметрии и трассировки операций очистки хвостов.
type TailCutInteractor[ID comparable] struct {
	*telemetry.BaseTelemetry
	spanName string
	next     extuc.TailCutUseCase[ID]
}

var _ extuc.TailCutUseCase[string] = (*TailCutInteractor[string])(nil)

// NewTailCutUseCase представляет собой фабричный метод для создания
// нового экземпляра трассирующего декоратора TailCutInteractor с поддержкой любых сопоставимых типов ID.
func NewTailCutUseCase[ID comparable](
	ucName string,
	next extuc.TailCutUseCase[ID],
) *TailCutInteractor[ID] {
	return &TailCutInteractor[ID]{
		next:          next,
		spanName:      fmt.Sprintf("%s.Cut", ucName),
		BaseTelemetry: telemetry.NewBaseTelemetry(ucName),
	}
}

// Cut выполняет декорируемый вызов для удаления элемента из хвоста по его идентификатору,
// фиксируя строковое представление ID в атрибутах спана, события ошибок и статус операции.
func (tci *TailCutInteractor[ID]) Cut(ctx context.Context, id ID) error {
	ucCtx, span := tci.StartSpan(ctx, tci.spanName)
	defer span.End()

	span.SetAttributes(
		attribute.String("param.id", fmt.Sprintf("%v", id)),
	)

	err := tci.next.Cut(ucCtx, id)
	if err != nil {
		span.AddEvent("Cut_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}
