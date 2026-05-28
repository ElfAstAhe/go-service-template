package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
)

// AppConfig — конфигурация приложения/сервиса
type AppConfig struct {
	*config.AppConfig `mapstructure:",squash"`
}

func NewAppConfig(
	env config.AppEnv,
	initTimeout time.Duration,
	stopTimeout time.Duration,
	closeTimeout time.Duration,
) *AppConfig {
	return &AppConfig{
		AppConfig: config.NewAppConfig(env, initTimeout, stopTimeout, closeTimeout),
	}
}

func NewDefaultAppConfig() *AppConfig {
	return NewAppConfig(
		config.DefaultAppEnv,
		config.DefaultAppInitTimeout,
		config.DefaultAppStopTimeout,
		config.DefaultAppCloseTimeout,
	)
}
