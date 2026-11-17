package config

import (
	"strings"

	pkgconf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkgextconf "github.com/ElfAstAhe/go-service-template/pkg/ext/config"
)

// App config flags
const (
	flagAppNodeName string = "app-node-name"
)

const (
	keyAppNodeName string = "app.node_name"
)

const (
	defaultNodeName string = "example-service"
)

// AppConfig — конфигурация приложения/сервиса
type AppConfig struct {
	*pkgconf.AppConfig `mapstructure:",squash"`
	NodeName           string `mapstructure:"node_name" json:"node_name,omitempty" yaml:"node_name,omitempty" toml:"node_name,omitempty"`
	MaxListLimit       int    `mapstructure:"max_list_limit" json:"max_list_limit,omitempty" yaml:"max_list_limit,omitempty" toml:"max_list_limit,omitempty"`
}

var _ pkgconf.Validator = (*AppConfig)(nil)

func NewAppConfig(
	base *pkgconf.AppConfig,
	nodeName string,
	maxListLimit int,
) *AppConfig {
	return &AppConfig{
		AppConfig:    base,
		NodeName:     nodeName,
		MaxListLimit: maxListLimit,
	}
}

func NewDefaultAppConfig() *AppConfig {
	return NewAppConfig(
		pkgconf.NewAppConfig(
			pkgconf.DefaultAppEnv,
			pkgconf.DefaultAppInitTimeout,
			pkgconf.DefaultAppStopTimeout,
			pkgconf.DefaultAppCloseTimeout,
		),
		defaultNodeName,
		pkgextconf.DefaultAppMaxListLimit,
	)
}

func (ac *AppConfig) Validate() error {
	if err := ac.AppConfig.Validate(); err != nil {
		return errs.NewConfigValidateError("app", "base", "validation failed", err)
	}
	if strings.TrimSpace(ac.NodeName) == "" {
		return errs.NewConfigValidateError("app", "node_name", "is required", nil)
	}
	if ac.MaxListLimit <= 0 {
		return errs.NewConfigValidateError("app", "max_list_limit", "is required", nil)
	}

	return nil
}

func BuildAppSettings() pkgconf.Settings {
	return append(
		pkgconf.BuildAppSettings(),
		pkgconf.NewSetting(keyAppNodeName, flagAppNodeName, defaultNodeName, "node name"),
		pkgconf.NewSetting(pkgextconf.KeyAppMaxListLimit, pkgextconf.FlagAppMaxListLimit, pkgextconf.DefaultAppMaxListLimit, "max list limit"),
	)
}
