package config

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// AuthConfig объединяет настройки всех поддерживаемых механизмов аутентификации сервиса
// (Basic Auth, JWT, клиентские TLS-сертификаты mTLS).
type AuthConfig struct {
	// Basic конфигурация HTTP/gRPC Basic-аутентификации
	Basic *AuthBasicConfig `mapstructure:"basic" json:"basic,omitempty" yaml:"basic,omitempty" toml:"basic,omitempty"`

	// JWT конфигурация выпуска и проверки JWT-токенов
	JWT *AuthJWTConfig `mapstructure:"jwt" json:"jwt,omitempty" yaml:"jwt,omitempty" toml:"jwt,omitempty"`

	// Cert конфигурация клиентской TLS-аутентификации (mTLS)
	Cert *AuthCertConfig `mapstructure:"cert" json:"cert,omitempty" yaml:"cert,omitempty" toml:"cert,omitempty"`
}

// NewAuthConfig — фабричный конструктор корневой конфигурации подсистемы аутентификации.
func NewAuthConfig(
	basic *AuthBasicConfig,
	jwt *AuthJWTConfig,
	cert *AuthCertConfig,
) *AuthConfig {
	return &AuthConfig{
		Basic: basic,
		JWT:   jwt,
		Cert:  cert,
	}
}

// NewDefaultAuthConfig собирает конфигурацию подсистемы аутентификации со значениями по умолчанию
// для всех поддерживаемых провайдеров.
func NewDefaultAuthConfig() *AuthConfig {
	return NewAuthConfig(
		NewDefaultAuthBasicConfig(),
		NewDefaultAuthJWTConfig(),
		NewDefaultAuthCertConfig(),
	)
}

// Validate осуществляет каскадную проверку всех инициализированных конфигураций провайдеров
// аутентификации на этапе запуска приложения (Fail-Fast).
func (ac *AuthConfig) Validate() error {
	if !utils.IsNil(ac.Basic) {
		if err := ac.Basic.Validate(); err != nil {
			return errs.NewConfigValidateError("auth", "basic", "validation failed", err)
		}
	}
	if !utils.IsNil(ac.JWT) {
		if err := ac.JWT.Validate(); err != nil {
			return errs.NewConfigValidateError("auth", "jwt", "validation failed", err)
		}
	}
	if !utils.IsNil(ac.Cert) {
		if err := ac.Cert.Validate(); err != nil {
			return errs.NewConfigValidateError("auth", "cert", "validation failed", err)
		}
	}

	return nil
}

// BuildDefaultAuthSettings формирует объединённый набор настроек Viper/PFlag
// для всех провайдеров аутентификации с соответствующими вложенными префиксами.
func BuildDefaultAuthSettings(keyPrefix, flagPrefix string) Settings {
	var res Settings
	res = append(res, BuildDefaultAuthBasicSettings(keyPrefix+".basic", flagPrefix)...)
	res = append(res, BuildDefaultAuthJWTSettings(keyPrefix+".jwt", flagPrefix)...)
	res = append(res, BuildDefaultAuthCertSettings(keyPrefix+".cert", flagPrefix)...)

	return res
}
