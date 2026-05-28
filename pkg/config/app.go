package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type AppConfig struct {
	Env          AppEnv        `mapstructure:"env" json:"env,omitempty" yaml:"env,omitempty"` // dev, prod, test
	InitTimeout  time.Duration `mapstructure:"init_timeout" json:"init_timeout,omitempty" yaml:"init_timeout,omitempty"`
	StopTimeout  time.Duration `mapstructure:"stop_timeout" json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty"`
	CloseTimeout time.Duration `mapstructure:"close_timeout" json:"close_timeout,omitempty" yaml:"close_timeout,omitempty"`
}

func NewAppConfig(
	env AppEnv,
	initTimeout time.Duration,
	stopTimeout time.Duration,
	closeTimeout time.Duration,
) *AppConfig {
	return &AppConfig{
		Env:          env,
		InitTimeout:  initTimeout,
		StopTimeout:  stopTimeout,
		CloseTimeout: closeTimeout,
	}
}

func NewDefaultAppConfig() *AppConfig {
	return NewAppConfig(
		DefaultAppEnv,
		DefaultAppInitTimeout,
		DefaultAppStopTimeout,
		DefaultAppCloseTimeout,
	)
}

func (ac *AppConfig) Validate() error {
	if ac.Env == "" {
		return errs.NewConfigValidateError("app", "Env", "empty", nil)
	}
	if !ac.Env.Exists() {
		return errs.NewConfigValidateError("app", "env", "env value not match", nil)
	}
	if ac.InitTimeout < 0 {
		return errs.NewConfigValidateError("app", "InitTimeout", "must be equal or greater zero", nil)
	}
	if ac.StopTimeout < 0 {
		return errs.NewConfigValidateError("app", "StopTimeout", "must be equal or greater zero", nil)
	}
	if ac.CloseTimeout < 0 {
		return errs.NewConfigValidateError("app", "CloseTimeout", "must be equal or greater zero", nil)
	}

	return nil
}
