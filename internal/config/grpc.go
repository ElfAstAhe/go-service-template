package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// GRPCConfig — настройки gRPC сервера
type GRPCConfig struct {
	Address     string        `mapstructure:"address"`
	MaxConnIdle time.Duration `mapstructure:"max_conn_idle"`
	MaxConnAge  time.Duration `mapstructure:"max_conn_age"`
	Timeout     time.Duration `mapstructure:"timeout"`
	// Настройки KeepAlive важны, чтобы соединения не "протухали" за балансировщиками
	KeepAliveTime    time.Duration `mapstructure:"keep_alive_time"`
	KeepAliveTimeout time.Duration `mapstructure:"keep_alive_timeout"`
}

func NewGRPCConfig(address string, maxConnIdle, maxConnAge, timeout, keepAliveTime, keepAliveTimeout time.Duration) *GRPCConfig {
	return &GRPCConfig{
		Address:          address,
		MaxConnIdle:      maxConnIdle,
		MaxConnAge:       maxConnAge,
		Timeout:          timeout,
		KeepAliveTime:    keepAliveTime,
		KeepAliveTimeout: keepAliveTimeout,
	}
}

func NewDefaultGRPCConfig() *GRPCConfig {
	return NewGRPCConfig(defaultGRPCAddress, 0, 0, defaultGRPCTimeout, 0, 0)
}

func (gc *GRPCConfig) Validate() error {
	if gc.Address == "" {
		return errs.NewConfigValidateError("gRPC", "address", "must not be empty", nil)
	}
	if gc.MaxConnIdle <= 0 {
		return errs.NewConfigValidateError("gRPC", "max_conn_idle", "must be greater than zero", nil)
	}
	if gc.MaxConnAge <= 0 {
		return errs.NewConfigValidateError("gRPC", "max_conn_age", "must be greater than zero", nil)
	}
	if gc.Timeout <= 0 {
		return errs.NewConfigValidateError("gRPC", "timeout", "must be greater than zero", nil)
	}
	if gc.KeepAliveTime <= 0 {
		return errs.NewConfigValidateError("gRPC", "keepalive_time", "must be greater than zero", nil)
	}
	if gc.KeepAliveTimeout <= 0 {
		return errs.NewConfigValidateError("gRPC", "keepalive_timeout", "must be greater than zero", nil)
	}

	return nil
}
