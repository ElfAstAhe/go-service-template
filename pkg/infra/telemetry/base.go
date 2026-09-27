package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// BaseTelemetry инкапсулирует базовые низкоуровневые механизмы работы с OpenTelemetry (OTel).
//
// Выступает в роли единой точки входа для управления распределенной трассировкой компонентов фреймворка.
// Спроектирован для сквозного встраивания (композиции) в декораторы репозиториев, клиентов и воркеров,
// изолируя прикладной код от рутины извлечения глобальных OTel-провайдеров.
type BaseTelemetry struct {
	name   string       // Уникальное строковое имя трасера (Tracer Name / Instrumentation Scope)
	tracer trace.Tracer // Экземпляр нативного OTel-трейсера текущего компонента
}

// NewBaseTelemetry — фабричный конструктор базового компонента телеметрии.
// Автоматически регистрирует трасер в глобальном реестре OpenTelemetry (otel.GetTracerProvider).
func NewBaseTelemetry(name string) *BaseTelemetry {
	return &BaseTelemetry{
		tracer: otel.GetTracerProvider().Tracer(name),
		name:   name,
	}
}

// StartSpan открывает новый именованный спан (Span) распределенной трассировки в рамках текущего контекста.
// Автоматически связывает родительские и дочерние спаны (Span Propagation) для визуализации в Jaeger/Loki.
func (bt *BaseTelemetry) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return bt.tracer.Start(ctx, name, opts...)
}

// GetTracer возвращает прямую ссылку на нативный OTel-интерфейс трейсера.
func (bt *BaseTelemetry) GetTracer() trace.Tracer {
	return bt.tracer
}

// GetTracerName возвращает строковый идентификатор области видимости текущего трассера.
func (bt *BaseTelemetry) GetTracerName() string {
	return bt.name
}
