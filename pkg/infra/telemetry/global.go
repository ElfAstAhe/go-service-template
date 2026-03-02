package telemetry

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
)

// SetupOTel конфигурирует глобальный TracerProvider и Propagator.
// Возвращает функцию shutdown, которую нужно вызвать при выходе из приложения.
func SetupOTel(ctx context.Context, cfg *config.TelemetryConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return nopTelemetryShutdown, nil
	}

	// 1. Экспортер (куда шлем данные)
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
		otlptracegrpc.WithInsecure(), // Для начала без TLS
	)
	if err != nil {
		return nil, err
	}

	// 2. Ресурс (описываем наш сервис)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// 3. TracerProvider (мозг системы)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter), // Пакетная отправка — это Senior-way
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRate))),
	)

	// Регистрируем глобально
	otel.SetTracerProvider(tp)

	// Важно: настраиваем проброс TraceID между сервисами (W3C Standard)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func nopTelemetryShutdown(ctx context.Context) error {
	return nil
}
