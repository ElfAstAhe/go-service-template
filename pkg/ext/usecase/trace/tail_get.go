package trace

import (
	"context"
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/ext/usecase"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// TailGetInteractor является универсальным (generic) декоратором для usecase.TailGetUseCase,
// предназначенным для автоматического сбора распределенной телеметрии и трассировки операций получения хвостов.
type TailGetInteractor[ID comparable] struct {
	*telemetry.BaseTelemetry
	spanName string
	next     usecase.TailGetUseCase[ID]
}

var _ usecase.TailGetUseCase[string] = (*TailGetInteractor[string])(nil)

// NewTailGetUseCase представляет собой фабричный метод для создания
// нового экземпляра трассирующего декоратора TailGetInteractor с поддержкой любых сопоставимых типов ID.
func NewTailGetUseCase[ID comparable](
	ucName string,
	next usecase.TailGetUseCase[ID],
) *TailGetInteractor[ID] {
	return &TailGetInteractor[ID]{
		BaseTelemetry: telemetry.NewBaseTelemetry(ucName),
		spanName:      fmt.Sprintf("%s.Get", ucName),
		next:          next,
	}
}

// Get выполняет декорируемый вызов для получения списка идентификаторов из хвоста, созданных до указанного времени,
// фиксируя метку времени в атрибутах спана, события ошибок и статус операции.
func (tgi *TailGetInteractor[ID]) Get(ctx context.Context, tail time.Time) ([]ID, error) {
	ucCtx, span := tgi.StartSpan(ctx, tgi.spanName)
	defer span.End()

	span.SetAttributes(
		attribute.Int64("param.tail", tail.Unix()),
	)

	res, err := tgi.next.Get(ucCtx, tail)
	if err != nil {
		span.AddEvent("Get_failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return res, err
}
