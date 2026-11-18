package config

import (
	"strings"

	pkgconf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// App config flags
//
//nolint:gosec // G101 : app params
const (
	FlagAppNodeName           string = "node-name"
	FlagAppMaxListLimit       string = "app-max-list-limit"
	FlagAppCipherKey          string = "app-cipher-key"
	FlagAppAcceptTokenIssuers string = "app-accept-token-issuers"
)

// app
//
//nolint:gosec // G101 : app params
const (
	KeyAppNodeName           string = "app.node_name"
	KeyAppMaxListLimit       string = "app.max_list_limit"
	KeyAppCipherKey          string = "app.cipher_key"
	KeyAppAcceptTokenIssuers string = "app.accept_token_issuers"
)

// app
//
//nolint:gosec // G101 : dev/test default conf vals
const (
	DefaultAppMaxListLimit int    = 100
	DefaultAppCipherKey    string = "123456"
)

// DefaultAcceptTokenIssuers default token issuers
var DefaultAcceptTokenIssuers []string

// AppConfig расширяет базовую конфигурацию фреймворка, инкапсулируя метаданные,
// лимиты выборок, параметры криптографии и правила валидации токенов безопасности микросервиса.
type AppConfig struct {
	*pkgconf.AppConfig `mapstructure:",squash"`
	NodeName           string   `mapstructure:"node_name" json:"node_name,omitempty" yaml:"node_name,omitempty" toml:"node_name,omitempty"`
	MaxListLimit       int      `mapstructure:"max_list_limit" json:"max_list_limit,omitempty" yaml:"max_list_limit,omitempty" toml:"max_list_limit,omitempty"`
	AcceptTokenIssuers []string `mapstructure:"accept_token_issuers" json:"accept_token_issuers,omitempty" yaml:"accept_token_issuers,omitempty" toml:"accept_token_issuers,omitempty"`
	CipherKey          string   `mapstructure:"cipher_key" json:"cipher_key,omitempty" yaml:"cipher_key,omitempty" toml:"cipher_key,omitempty"`
}

// NewAppConfig — параметризованный конструктор для явного формирования конфигурационной структуры.
func NewAppConfig(
	base *pkgconf.AppConfig,
	nodeName string,
	maxListLimit int,
	acceptTokenIssuers []string,
	cipherKey string,
) *AppConfig {
	return &AppConfig{
		AppConfig:          base,
		NodeName:           nodeName,
		MaxListLimit:       maxListLimit,
		AcceptTokenIssuers: acceptTokenIssuers,
		CipherKey:          cipherKey,
	}
}

// NewDefaultAppConfig создает дефолтный профиль конфигурации, использующий константы базового шаблона.
func NewDefaultAppConfig(
	env string,
	nodeName string,
) *AppConfig {
	return NewAppConfig(
		pkgconf.NewDefaultAppConfig(env),
		nodeName,
		DefaultAppMaxListLimit,
		DefaultAcceptTokenIssuers,
		DefaultAppCipherKey,
	)
}

// Validate проверяет целостность и бизнес-валидность загруженных параметров конфигурации на фазе холодного старта.
func (ac *AppConfig) Validate() error {
	if err := ac.AppConfig.Validate(); err != nil {
		return errs.NewConfigValidateError("app", "app config", "validation failed", err)
	}
	if strings.TrimSpace(ac.NodeName) == "" {
		return errs.NewConfigValidateError("app", "node_name", "is required", nil)
	}
	if ac.MaxListLimit <= 0 {
		return errs.NewConfigValidateError("app", "max_list_limit", "must be greater zero", nil)
	}

	if ac.CipherKey == "" {
		return errs.NewConfigValidateError("app", "cipher_key", "must not be empty", nil)
	}

	// ToDo: restore in future
	//if len(ac.AcceptTokenIssuers) == 0 {
	//	return errs.NewConfigValidateError("app", "accept-token-issuers", "must not be empty", nil)
	//}

	return nil
}

// BuildAppSettings app settings builder
func BuildAppSettings(
	nodeName string,
) pkgconf.Settings {
	return append(pkgconf.BuildAppSettings(),
		pkgconf.NewSetting(KeyAppNodeName, FlagAppNodeName, nodeName, "application node name"),
		pkgconf.NewSetting(KeyAppMaxListLimit, FlagAppMaxListLimit, nodeName, "max list limit"),
		pkgconf.NewSetting(KeyAppCipherKey, FlagAppCipherKey, nodeName, "cipher key"),
		pkgconf.NewSetting(KeyAppAcceptTokenIssuers, FlagAppAcceptTokenIssuers, DefaultAcceptTokenIssuers, `accept token issuers separated by comma, like: "issuer1,issuer2,issuer3"`),
	)
}
