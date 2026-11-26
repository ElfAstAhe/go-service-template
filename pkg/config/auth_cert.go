package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги подсистемы клиентской TLS-аутентификации (mTLS / Cert Auth).
//
//nolint:gosec // G101 : это флаги (параметры)
const (
	TemplateFlagAuthCertEnabled    string = "%s-auth-cert"             // CLI-флаг включения аутентификации по клиентским сертификатам
	TemplateFlagAuthCertCAPath     string = "%s-auth-cert-ca-path"     // CLI-флаг пути к корневому CA-сертификату проверки клиентов
	TemplateFlagAuthCertAllowedCNs string = "%s-auth-cert-allowed-cns" // CLI-флаг списка разрешённых Common Name (CN)
	TemplateFlagAuthCertRequired   string = "%s-auth-cert-required"    // CLI-флаг обязательности предъявления сертификата на уровне TLS
)

// Строковые ключи Viper-маппинга для иерархического связывания параметров mTLS.
//
//nolint:gosec // G101 : это ключи
const (
	TemplateKeyAuthCertEnabled    string = "%s.enabled"
	TemplateKeyAuthCertCAPath     string = "%s.ca_path"
	TemplateKeyAuthCertAllowedCNs string = "%s.allowed_cns"
	TemplateKeyAuthCertRequired   string = "%s.required"
)

// Значения по умолчанию для клиентской TLS-аутентификации.
const (
	DefaultAuthCertEnabled  bool = false
	DefaultAuthCertRequired bool = true
)

// AuthCertConfig инкапсулирует параметры проверки клиентских X.509-сертификатов (mTLS)
// для подсистемы безопасности (IAM/Auth).
type AuthCertConfig struct {
	// Enabled признак использования аутентификации по клиентским сертификатам
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty" yaml:"enabled,omitempty" toml:"enabled,omitempty"`

	// CAPath путь в файловой системе к корневому сертификату CA для верификации цепочки доверия клиента
	CAPath string `mapstructure:"ca_path" json:"ca_path,omitempty" yaml:"ca_path,omitempty" toml:"ca_path,omitempty"`

	// AllowedCNs белый список разрешённых Common Name (CN) субъектов сертификата (если пуст — разрешены все валидные от CA)
	AllowedCNs []string `mapstructure:"allowed_cns" json:"allowed_cns,omitempty" yaml:"allowed_cns,omitempty" toml:"allowed_cns,omitempty"`

	// Required определяет строгий режим TLS (RequireAndVerifyClientCert при true, VerifyClientCertIfGiven при false)
	Required bool `mapstructure:"required" json:"required,omitempty" yaml:"required,omitempty" toml:"required,omitempty"`
}

// NewAuthCertConfig — фабричный конструктор конфигурации клиентской TLS-аутентификации.
func NewAuthCertConfig(
	enabled bool,
	caPath string,
	allowedCNs []string,
	required bool,
) *AuthCertConfig {
	return &AuthCertConfig{
		Enabled:    enabled,
		CAPath:     caPath,
		AllowedCNs: allowedCNs,
		Required:   required,
	}
}

// NewDefaultAuthCertConfig собирает конфигурацию mTLS по умолчанию.
func NewDefaultAuthCertConfig() *AuthCertConfig {
	return NewAuthCertConfig(
		DefaultAuthCertEnabled,
		"",
		nil,
		DefaultAuthCertRequired,
	)
}

// Validate осуществляет семантическую и системную проверку путей к CA-сертификату на этапе запуска (Fail-Fast).
func (acc *AuthCertConfig) Validate() error {
	if !acc.Enabled {
		return nil
	}
	if strings.TrimSpace(acc.CAPath) == "" {
		return errs.NewConfigValidateError("cert auth", "CAPath", "is required", nil)
	}
	if _, err := os.Stat(acc.CAPath); err != nil {
		if os.IsNotExist(err) {
			return errs.NewConfigValidateError("cert auth", "CAPath", "does not exist", err)
		}

		return errs.NewConfigValidateError("cert auth", "CAPath", "path check failed", err)
	}

	return nil
}

// BuildDefaultAuthCertSettings формирует набор настроек Viper/PFlag для AuthCertConfig.
//
// # ВНИМАНИЕ!!!
//
// Сюда не входит:
//
//	TemplateKeyAuthCertCAPath TemplateFlagAuthCertCAPath
//	TemplateKeyAuthCertAllowedCNs TemplateFlagAuthCertAllowedCNs
func BuildDefaultAuthCertSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeyAuthCertEnabled, keyPrefix), fmt.Sprintf(TemplateFlagAuthCertEnabled, flagPrefix), DefaultAuthCertEnabled, fmt.Sprintf("%s use cert auth", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAuthCertRequired, keyPrefix), fmt.Sprintf(TemplateFlagAuthCertRequired, flagPrefix), DefaultAuthCertRequired, fmt.Sprintf("%s require client certificate", flagPrefix)),
	}
}
