package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type BaseTelemetry struct {
	name   string
	tracer trace.Tracer
}

func NewBaseTelemetry(name string) *BaseTelemetry {
	return &BaseTelemetry{
		tracer: otel.GetTracerProvider().Tracer(name),
		name:   name,
	}
}

func (bt *BaseTelemetry) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return bt.tracer.Start(ctx, name, opts...)
}

func (bt *BaseTelemetry) GetTracer() trace.Tracer {
	return bt.tracer
}

func (bt *BaseTelemetry) GetTracerName() string {
	return bt.name
}
