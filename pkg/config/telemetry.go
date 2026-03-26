package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type TelemetryConfig struct {
	Enabled          bool          `mapstructure:"enabled" json:"enabled,omitempty" yaml:"enabled,omitempty"`
	ExporterEndpoint string        `mapstructure:"exporter_endpoint" json:"exporter_endpoint,omitempty" yaml:"exporter_endpoint,omitempty"` // например, "localhost:4317" для Jaeger/Tempo
	ServiceName      string        `mapstructure:"service_name" json:"service_name,omitempty" yaml:"service_name,omitempty"`
	SampleRate       float64       `mapstructure:"sample_rate" json:"sample_rate,omitempty" yaml:"sample_rate,omitempty"` // от 0.0 до 1.0
	Timeout          time.Duration `mapstructure:"timeout" json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func NewTelemetryConfig(enabled bool, exporterEndpoint string, serviceName string, sampleRate float64, timeout time.Duration) *TelemetryConfig {
	return &TelemetryConfig{
		Enabled:          enabled,
		ExporterEndpoint: exporterEndpoint,
		ServiceName:      serviceName,
		SampleRate:       sampleRate,
		Timeout:          timeout,
	}
}

func NewDefaultTelemetryConfig() *TelemetryConfig {
	return NewTelemetryConfig(DefaultTelemetryEnabled, DefaultTelemetryExporterEndpoint, "", DefaultTelemetrySampleRate, DefaultTelemetryTimeout)
}

func (tc *TelemetryConfig) Validate() error {
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
