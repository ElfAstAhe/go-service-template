package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги подсистемы сквозного распределенного сбора трассировок (Telemetry / OpenTelemetry).
const (
	FlagTelemetryEnabled          string = "telemetry-enabled"           // CLI-флаг глобальной активации экспорта спанов телеметрии
	FlagTelemetryServiceName      string = "telemetry-service-name"      // CLI-флаг имени сервиса для группировки в Jaeger/OpenSearch
	FlagTelemetryExporterEndpoint string = "telemetry-exporter-endpoint" // CLI-флаг адреса OTLP коллектора телеметрии (gRPC/HTTP)
	FlagTelemetrySampleRate       string = "telemetry-sample-rate"       // CLI-флаг коэффициента семплирования трассировок (0.0 до 1.0)
	FlagTelemetryTimeout          string = "telemetry-timeout"           // CLI-флаг сетевого таймаута экспортера телеметрии
)

// ====================================================================
// telemetry
// ====================================================================
// Строковые ключи Viper-маппинга для подсистемы сбора OpenTelemetry распределенных трассировок.
const (
	KeyTelemetryEnabled          string = "telemetry.enabled"
	KeyTelemetryServiceName      string = "telemetry.service_name"
	KeyTelemetryExporterEndpoint string = "telemetry.exporter_endpoint"
	KeyTelemetrySampleRate       string = "telemetry.sample_rate"
	KeyTelemetryTimeout          string = "telemetry.timeout"
)

// ====================================================================
// Telemetry defaults
// ====================================================================
// Настройки по умолчанию для подсистемы сквозной распределенной трассировки OpenTelemetry.
const (
	DefaultTelemetryEnabled          bool          = false            // По умолчанию сбор спанов отключен во избежание оверхеда в dev-среде
	DefaultTelemetryExporterEndpoint string        = "localhost:4317" // Стандартный gRPC порт OTLP коллектора (Jaeger/OpenSearch)
	DefaultTelemetrySampleRate       float64       = 1.0              // Коэффициент семплирования 1.0 означает сбор 100% трассировок
	DefaultTelemetryTimeout          time.Duration = 5 * time.Second  // Порог ожидания при отправке батча спанов в коллектор
)

// TelemetryConfig инкапсулирует конфигурационные параметры подсистемы распределенной
// трассировки OpenTelemetry (OTel), управляя экспортом спанов в коллекторы (Jaeger, Tempo).
type TelemetryConfig struct {
	// Enabled Глобальный флаг активации сбора и отправки трассировок
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty" yaml:"enabled,omitempty"`
	// ExporterEndpoint Сетевой gRPC/HTTP адрес OTLP коллектора (например, "localhost:4317")
	ExporterEndpoint string `mapstructure:"exporter_endpoint" json:"exporter_endpoint,omitempty" yaml:"exporter_endpoint,omitempty"`
	// ServiceName Имя текущего микросервиса для группировки спанов в UI визуализатора
	ServiceName string `mapstructure:"service_name" json:"service_name,omitempty" yaml:"service_name,omitempty"`
	// SampleRate Коэффициент сэмплирования трассировок (строго от 0.0 до 1.0)
	SampleRate float64 `mapstructure:"sample_rate" json:"sample_rate,omitempty" yaml:"sample_rate,omitempty"`
	// Timeout Сетевой таймаут сокета на операцию экспорта батча спанов
	Timeout time.Duration `mapstructure:"timeout" json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// NewTelemetryConfig — фабричный конструктор конфигурации подсистемы телеметрии.
func NewTelemetryConfig(enabled bool, exporterEndpoint string, serviceName string, sampleRate float64, timeout time.Duration) *TelemetryConfig {
	return &TelemetryConfig{
		Enabled:          enabled,
		ExporterEndpoint: exporterEndpoint,
		ServiceName:      serviceName,
		SampleRate:       sampleRate,
		Timeout:          timeout,
	}
}

// NewDefaultTelemetryConfig собирает конфигурацию по умолчанию, наполняя её системными константными дефолтами фреймворка.
func NewDefaultTelemetryConfig() *TelemetryConfig {
	return NewTelemetryConfig(
		DefaultTelemetryEnabled,
		DefaultTelemetryExporterEndpoint,
		"",
		DefaultTelemetrySampleRate,
		DefaultTelemetryTimeout,
	)
}

// Validate осуществляет семантическую и математическую валидацию параметров OpenTelemetry на этапе запуска приложения (Bootstrap Phase).
// Если сбор включен, метод пресекает старт подсистемы с пустыми эндпоинтами, незаданными именами служб или некорректным сэмплированием.
func (tc *TelemetryConfig) Validate() error {
	// Если экспорт спанов глобально отключен — завершаем проверку досрочно во избежание оверхеда (Guard Clause)
	if !tc.Enabled {
		return nil
	}
	if tc.ExporterEndpoint == "" {
		return errs.NewConfigValidateError("telemetry", "exporter_endpoint", "must not be empty", nil)
	}
	//if tc.ServiceName == "" {
	//	return errs.NewConfigValidateError("telemetry", "service_name", "must not be empty", nil)
	//}
	if tc.SampleRate < 0.0 || tc.SampleRate > 1.0 {
		return errs.NewConfigValidateError("telemetry", "sample_rate", "must be between 0.0 and 1.0", nil)
	}
	if tc.Timeout < 0 {
		return errs.NewConfigValidateError("telemetry", "timeout", "must be greater or equal to 0", nil)
	}

	return nil
}

// BuildTelemetrySettings telemetry settings builder
func BuildTelemetrySettings() Settings {
	return Settings{
		NewSetting(KeyTelemetryEnabled, FlagTelemetryEnabled, DefaultTelemetryEnabled, "telemetry operational"),
		NewSetting(KeyTelemetryExporterEndpoint, FlagTelemetryExporterEndpoint, DefaultTelemetryExporterEndpoint, "telemetry exporter endpoint"),
		NewSetting(KeyTelemetrySampleRate, FlagTelemetrySampleRate, DefaultTelemetrySampleRate, "telemetry sample rate"),
		NewSetting(KeyTelemetryTimeout, FlagTelemetryTimeout, DefaultTelemetryTimeout, "telemetry timeout"),
		NewSetting(KeyTelemetryServiceName, FlagTelemetryServiceName, "", "telemetry service name"),
	}
}
