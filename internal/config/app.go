package config

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AppConfig — метаданные сервиса
type AppConfig struct {
	Env AppEnv `mapstructure:"env" json:"env,omitempty" yaml:"env,omitempty"` // dev, prod, test
}

func NewAppConfig(env AppEnv) *AppConfig {
	return &AppConfig{
		Env: env,
	}
}

func NewDefaultAppConfig() *AppConfig {
	return NewAppConfig(defaultAppEnv)
}

func (ac *AppConfig) Validate() error {
	if ac.Env == "" {
		return errs.NewConfigValidateError("app", "env", "must not be empty", nil)
	}

	if !ac.Env.Exists() {
		return errs.NewConfigValidateError("app", "env", "env value not match", nil)
	}

	return nil
}
