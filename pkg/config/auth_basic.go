package config

import (
	"fmt"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги подсистемы Basic-аутентификации.
//
//nolint:gosec // G101 : это флаги (параметры)
const (
	TemplateFlagAuthBasicEnabled string = "%s-auth-basic"       // CLI-флаг включения Basic-аутентификации
	TemplateFlagAuthBasicRealm   string = "%s-auth-basic-realm" // CLI-флаг наименования защищаемой зоны (Realm)
)

// Строковые ключи Viper-маппинга для иерархического связывания параметров Basic Auth.
//
//nolint:gosec // G101 : это ключи
const (
	TemplateKeyAuthBasicEnabled string = "%s.enabled"
	TemplateKeyAuthBasicRealm   string = "%s.realm"
	TemplateKeyAuthBasicUsers   string = "%s.users"
)

// Значения по умолчанию для Basic-аутентификации.
const (
	DefaultAuthBasicEnabled bool   = false
	DefaultAuthBasicRealm   string = "Restricted"
)

// AuthBasicConfig инкапсулирует параметры HTTP/gRPC Basic-аутентификации
// для подсистемы безопасности (IAM/Auth).
type AuthBasicConfig struct {
	// Enabled признак использования Basic-аутентификации
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty" yaml:"enabled,omitempty" toml:"enabled,omitempty"`

	// Realm строковый идентификатор защищаемой области безопасности (заголовок WWW-Authenticate)
	Realm string `mapstructure:"realm" json:"realm,omitempty" yaml:"realm,omitempty" toml:"realm,omitempty"`

	// Users словарь учетных записей (login -> password или хэш пароля)
	Users map[string]string `mapstructure:"users" json:"users,omitempty" yaml:"users,omitempty" toml:"users,omitempty"`
}

// NewAuthBasicConfig — фабричный конструктор конфигурации Basic-аутентификации.
func NewAuthBasicConfig(
	enabled bool,
	realm string,
	users map[string]string,
) *AuthBasicConfig {
	return &AuthBasicConfig{
		Enabled: enabled,
		Realm:   realm,
		Users:   users,
	}
}

// NewDefaultAuthBasicConfig собирает конфигурацию Basic-аутентификации по умолчанию.
func NewDefaultAuthBasicConfig() *AuthBasicConfig {
	return NewAuthBasicConfig(
		DefaultAuthBasicEnabled,
		DefaultAuthBasicRealm,
		make(map[string]string),
	)
}

// Validate осуществляет семантическую проверку параметров Basic-аутентификации на этапе запуска (Fail-Fast).
func (abc *AuthBasicConfig) Validate() error {
	if !abc.Enabled {
		return nil
	}
	if strings.TrimSpace(abc.Realm) == "" {
		return errs.NewConfigValidateError("basic auth", "Realm", "is required", nil)
	}
	if len(abc.Users) == 0 {
		return errs.NewConfigValidateError("basic auth", "Users", "must contain at least one user", nil)
	}
	for login, pass := range abc.Users {
		if strings.TrimSpace(login) == "" {
			return errs.NewConfigValidateError("basic auth", "Users", "login must not be empty", nil)
		}
		if strings.TrimSpace(pass) == "" {
			return errs.NewConfigValidateError("basic auth", "Users", fmt.Sprintf("password for user [%s] must not be empty", login), nil)
		}
	}

	return nil
}

// BuildDefaultAuthBasicSettings формирует набор настроек Viper/PFlag для AuthBasicConfig.
//
// # ВНИМАНИЕ!!!
//
// Сюда не входит:
//
// TemplateKeyAuthBasicUsers
func BuildDefaultAuthBasicSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeyAuthBasicEnabled, keyPrefix), fmt.Sprintf(TemplateFlagAuthBasicEnabled, flagPrefix), DefaultAuthBasicEnabled, fmt.Sprintf("%s use basic auth", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAuthBasicRealm, keyPrefix), fmt.Sprintf(TemplateFlagAuthBasicRealm, flagPrefix), DefaultAuthBasicRealm, fmt.Sprintf("%s basic auth realm", flagPrefix)),
	}
}
