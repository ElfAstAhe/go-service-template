package telemetry

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
)

// SetupOTel производит сквозную инициализацию OpenTelemetry SDK: настраивает gRPC-экспортер,
// описывает метаданные сервиса, конфигурирует стратегии сэмплинга и регистрирует глобальные провайдеры.
//
// Возвращает функцию-заглушку shutdown для дефер-вызова в main.go, гарантирующую Flush накопленных спанов на диск/сеть.
func SetupOTel(ctx context.Context, cfg *config.TelemetryConfig) (func(context.Context) error, error) {
	// Если сбор телеметрии принудительно отключен в конфигурации (например, на локальном ПК разработчика)
	if !cfg.Enabled {
		return nopTelemetryShutdown, nil
	}

	// 1. Конфигурируем сетевой OTLP gRPC Экспортер (куда отправляем собранные трейсы: Jaeger, OpenSearch, Tempo)
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
		otlptracegrpc.WithInsecure(), // Отключаем TLS на этапе отладки внутри закрытого периметра K8s
	)
	if err != nil {
		return nil, errs.NewCommonError("create trace exporter failed", err)
	}

	// 2. Создаем Ресурс (Resource) — описываем метаданные и семантические атрибуты текущего пода/сервиса
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName), // Название микросервиса (например, "tiny-audit")
		),
	)
	if err != nil {
		return nil, errs.NewCommonError("create trace resource description failed", err)
	}

	// 3. Собираем ядро TracerProvider — мозг распределенной трассировки OpenTelemetry
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter), // Пакетная асинхронная отправка спанов (Senior-way) исключает тормоза сокетов
		sdktrace.WithResource(res),
		// Умный сэмплинг: если родительский запрос уже сэмплирован — пишем его трейс целиком.
		// Если это холодный старт — берем случайную выборку на основе переданного коэффициента Ratio (0.0 - 1.0).
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRate))),
	)

	// Регистрируем собранный провайдер на глобальном уровне экосистемы Go
	otel.SetTracerProvider(tp)

	// Настраиваем сквозной проброс метаданных TraceID и Baggage между сервисами по стандарту W3C Standard
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // Отвечает за бесшовную передачу заголовков traceparent
		propagation.Baggage{},      // Отвечает за трансляцию кастомного пользовательского контекста
	))

	return tp.Shutdown, nil
}

// nopTelemetryShutdown — пустая функция-заглушка (No-op), используемая при отключенной телеметрии.
func nopTelemetryShutdown(ctx context.Context) error {
	return nil
}
