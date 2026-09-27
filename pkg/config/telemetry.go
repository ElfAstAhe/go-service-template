package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// TelemetryConfig инкапсулирует конфигурационные параметры подсистемы распределенной
// трассировки OpenTelemetry (OTel), управляя экспортом спанов в коллекторы (Jaeger, Tempo).
type TelemetryConfig struct {
	Enabled          bool          `mapstructure:"enabled" json:"enabled,omitempty" yaml:"enabled,omitempty"`                               // Глобальный флаг активации сбора и отправки трассировок
	ExporterEndpoint string        `mapstructure:"exporter_endpoint" json:"exporter_endpoint,omitempty" yaml:"exporter_endpoint,omitempty"` // Сетевой gRPC/HTTP адрес OTLP коллектора (например, "localhost:4317")
	ServiceName      string        `mapstructure:"service_name" json:"service_name,omitempty" yaml:"service_name,omitempty"`                // Имя текущего микросервиса для группировки спанов в UI визуализатора
	SampleRate       float64       `mapstructure:"sample_rate" json:"sample_rate,omitempty" yaml:"sample_rate,omitempty"`                   // Коэффициент сэмплирования трассировок (строго от 0.0 до 1.0)
	Timeout          time.Duration `mapstructure:"timeout" json:"timeout,omitempty" yaml:"timeout,omitempty"`                               // Сетевой таймаут сокета на операцию экспорта батча спанов
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
	return NewTelemetryConfig(DefaultTelemetryEnabled, DefaultTelemetryExporterEndpoint, "", DefaultTelemetrySampleRate, DefaultTelemetryTimeout)
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
	if tc.ServiceName == "" {
		return errs.NewConfigValidateError("telemetry", "service_name", "must not be empty", nil)
	}
	if tc.SampleRate < 0.0 || tc.SampleRate > 1.0 {
		return errs.NewConfigValidateError("telemetry", "sample_rate", "must be between 0.0 and 1.0", nil)
	}
	if tc.Timeout < 0 {
		return errs.NewConfigValidateError("telemetry", "timeout", "must be greater or equal to 0", nil)
	}

	return nil
}
