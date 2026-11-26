package config

import (
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	TemplateKeyBasicCredsUsername string = "%s.username"
	TemplateKeyBasicCredsPassword string = "%s.password"
)

// BasicCredentialConfig содержит параметры конфигурации базовых учетных данных
// (имя пользователя и пароль) для аутентификации в клиентских подключениях.
type BasicCredentialConfig struct {
	Username string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty" toml:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`
}

// NewBasicCredentialConfig создает и инициализирует новый экземпляр конфигурации BasicCredentialConfig.
func NewBasicCredentialConfig(username string, password string) *BasicCredentialConfig {
	return &BasicCredentialConfig{
		Username: username,
		Password: password,
	}
}

// NewDefaultBasicCredentialConfig создает экземпляр конфигурации BasicCredentialConfig со значениями по умолчанию.
func NewDefaultBasicCredentialConfig() *BasicCredentialConfig {
	return NewBasicCredentialConfig("", "")
}

// Validate проверяет корректность заполнения имени пользователя и пароля в конфигурации BasicCredentialConfig.
func (bc *BasicCredentialConfig) Validate() error {
	if strings.TrimSpace(bc.Username) == "" {
		return errs.NewConfigValidateError("basic credential", "username", "must not be empty", nil)
	}
	if strings.TrimSpace(bc.Password) == "" {
		return errs.NewConfigValidateError("basic credential", "password", "must not be empty", nil)
	}

	return nil
}
