package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// GRPCConfig — настройки gRPC сервера
type GRPCConfig struct {
	Address         string        `mapstructure:"address" json:"address,omitempty" yaml:"address,omitempty"`
	MaxConnIdle     time.Duration `mapstructure:"max_conn_idle" json:"max_conn_idle,omitempty" yaml:"max_conn_idle,omitempty"`
	MaxConnAge      time.Duration `mapstructure:"max_conn_age" json:"max_conn_age,omitempty" yaml:"max_conn_age,omitempty"`
	MaxConnAgeGrace time.Duration `mapstructure:"max_conn_age_grace" json:"max_conn_age_grace,omitempty" yaml:"max_conn_age_grace,omitempty"`
	Timeout         time.Duration `mapstructure:"timeout" json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Настройки KeepAlive важны, чтобы соединения не "протухали" за балансировщиками
	KeepAliveTime    time.Duration `mapstructure:"keep_alive_time" json:"keep_alive_time,omitempty" yaml:"keep_alive_time,omitempty"`
	KeepAliveTimeout time.Duration `mapstructure:"keep_alive_timeout" json:"keep_alive_timeout,omitempty" yaml:"keep_alive_timeout,omitempty"`
	ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty"`
}

func NewGRPCConfig(
	address string,
	maxConnIdle,
	maxConnAge,
	maxConnAgeGrace,
	timeout,
	keepAliveTime,
	keepAliveTimeout,
	shutdownTimeout time.Duration,
) *GRPCConfig {
	return &GRPCConfig{
		Address:          address,
		MaxConnIdle:      maxConnIdle,
		MaxConnAge:       maxConnAge,
		MaxConnAgeGrace:  maxConnAgeGrace,
		Timeout:          timeout,
		KeepAliveTime:    keepAliveTime,
		KeepAliveTimeout: keepAliveTimeout,
		ShutdownTimeout:  shutdownTimeout,
	}
}

func NewDefaultGRPCConfig() *GRPCConfig {
	return NewGRPCConfig(
		DefaultGRPCAddress,
		DefaultGRPCMaxConnIdle,
		DefaultGRPCMaxConnAge,
		DefaultGRPCMaxConnAgeGrace,
		DefaultGRPCTimeout,
		DefaultGRPCKeepAliveTime,
		DefaultGRPCKeepAliveTimeout,
		DefaultGRPCShutdownTimeout)
}

func (gc *GRPCConfig) Validate() error {
	if gc.Address == "" {
		return errs.NewConfigValidateError("gRPC", "address", "must not be empty", nil)
	}
	if gc.MaxConnIdle < 0 {
		return errs.NewConfigValidateError("gRPC", "max_conn_idle", "must be greater or equal zero", nil)
	}
	if gc.MaxConnAge < 0 {
		return errs.NewConfigValidateError("gRPC", "max_conn_age", "must be greater or equal zero", nil)
	}
	if gc.MaxConnAgeGrace < 0 {
		return errs.NewConfigValidateError("gRPC", "max_conn_age_grace", "must be greater or equal zero", nil)
	}
	if gc.Timeout < 0 {
		return errs.NewConfigValidateError("gRPC", "timeout", "must be greater or equal zero", nil)
	}
	if gc.KeepAliveTime < 0 {
		return errs.NewConfigValidateError("gRPC", "keepalive_time", "must be greater or equal zero", nil)
	}
	if gc.KeepAliveTimeout < 0 {
		return errs.NewConfigValidateError("gRPC", "keepalive_timeout", "must be greater or equal zero", nil)
	}
	if gc.ShutdownTimeout < 0 {
		return errs.NewConfigValidateError("gRPC", "shutdown_timeout", "must be greater or equal zero", nil)
	}

	return nil
}
