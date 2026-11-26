package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/helper/jwt"
)

// CLI-флаги подсистемы безопасности, шифрования и авторизации (IAM/Auth).
//
//nolint:gosec // G101 : это флаги (параметры)
const (
	TemplateFlagAuthJWTEnabled           string = "%s-auth-jwt"
	TemplateFlagAuthJWTSecret            string = "%s-auth-jwt-secret"         // CLI-флаг секретной соли HMAC подписи JWT токенов
	TemplateFlagAuthJWTSigningMethod     string = "%s-auth-jwt-signing-method" // CLI-флаг алгоритма подписи JWT (например, HS256)
	TemplateFlagAuthJWTIssuer            string = "%s-auth-jwt-issuer"
	TemplateFlagAuthJWTLeeway            string = "%s-auth-jwt-leeway"               // CLI-флаг времени погрешности валидации токена
	TemplateFlagAuthJWTAccessTokenTTL    string = "%s-auth-jwt-access-token-ttl"     // CLI-флаг времени жизни Access-токена
	TemplateFlagAuthJWTRefreshTokenTTL   string = "%s-auth-jwt-refresh-token-ttl"    // CLI-флаг времени жизни Refresh-токена
	TemplateFlagAuthJWTRSAPrivateKeyPath string = "%s-auth-jwt-rsa-private-key-path" // CLI-флаг пути к приватному RSA-ключу на диске
)

// Строковые ключи Viper-маппинга для иерархического связывания секретов подсистемы безопасности и IAM.
//
//nolint:gosec // G101 : это ключи
const (
	TemplateKeyAuthJWTEnabled           string = "%s.enabled"
	TemplateKeyAuthJWTSecret            string = "%s.secret"
	TemplateKeyAuthJWTSigningMethod     string = "%s.signing_method"
	TemplateKeyAuthJWTIssuer            string = "%s.issuer"
	TemplateKeyAuthJWTLeeway            string = "%s.leeway"
	TemplateKeyAuthJWTAccessTokenTTL    string = "%s.access_token_ttl"
	TemplateKeyAuthJWTRefreshTokenTTL   string = "%s.refresh_token_ttl"
	TemplateKeyAuthJWTRSAPrivateKeyPath string = "%s.rsa_private_key_path"
)

// Константные таймауты авторизации и параметры безопасности сессий по умолчанию.
const (
	DefaultAuthJWTEnabled         bool          = true
	DefaultAuthJWTSecret          string        = "123456"
	DefaultAuthJWTSigningMethod   string        = "HS256" // Метод подписи токенов по умолчанию: HMAC-SHA256
	DefaultAuthJWTLeeway          time.Duration = 30 * time.Second
	DefaultAuthJWTAccessTokenTTL  time.Duration = 15 * time.Minute // Время жизни короткоживущего токена доступа (Access Token)
	DefaultAuthJWTRefreshTokenTTL time.Duration = 24 * time.Hour   // Время жизни сессионного токена обновления (Refresh Token)
)

// AuthJWTConfig инкапсулирует криптографические секреты, алгоритмы подписи,
// временные лимиты сессий (JWT TTL) и параметры хэширования для подсистемы безопасности (IAM/Auth).
type AuthJWTConfig struct {
	// Enabled признак использования JWT
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty" yaml:"enabled,omitempty" toml:"enabled,omitempty"`

	// Secret секретный ключ (соль) для симметричной HMAC-подписи токенов
	Secret string `mapstructure:"secret" json:"secret,omitempty" yaml:"secret,omitempty" toml:"secret,omitempty"`

	// SigningMethod строковое имя алгоритма подписи (например, "HS256", "RS256")
	SigningMethod string `mapstructure:"signing_method" json:"signing_method,omitempty" yaml:"signing_method,omitempty" toml:"signing_method,omitempty"`

	// Issuer идентификатор выпускабщего центра/сервиса
	Issuer string `mapstructure:"issuer" json:"issuer,omitempty" yaml:"issuer,omitempty" toml:"issuer,omitempty"`

	// Leeway интервал времени погрешности валидации токена
	Leeway time.Duration `mapstructure:"leeway" json:"leeway,omitempty" yaml:"leeway,omitempty" toml:"leeway,omitempty"`

	// AccessTokenTTL время жизни Access-токена (Default: 15m)
	AccessTokenTTL time.Duration `mapstructure:"access_token_ttl" json:"access_token_ttl,omitempty" yaml:"access_token_ttl,omitempty" toml:"access_token_ttl,omitempty"`

	// RefreshTokenTTL время жизни Refresh-токена (Default: 30d)
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" json:"refresh_token_ttl,omitempty" yaml:"refresh_token_ttl,omitempty" toml:"refresh_token_ttl,omitempty"`

	// RSAPrivateKeyPath путь в файловой системе к приватному RSA-ключу для асимметричной подписи токенов
	RSAPrivateKeyPath string `mapstructure:"rsa_private_key_path" json:"rsa_private_key_path,omitempty" yaml:"rsa_private_key_path,omitempty" toml:"rsa_private_key_path,omitempty"`
}

// NewAuthJWTConfig — фабричный конструктор конфигурации подсистемы безопасности.
func NewAuthJWTConfig(
	enabled bool,
	secret string,
	signingMethod string,
	issuer string,
	leeway time.Duration,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	rsaPrivateKeyPath string,
) *AuthJWTConfig {
	return &AuthJWTConfig{
		Enabled:           enabled,
		Secret:            secret,
		SigningMethod:     signingMethod,
		Issuer:            issuer,
		Leeway:            leeway,
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
		RSAPrivateKeyPath: rsaPrivateKeyPath,
	}
}

// NewDefaultAuthJWTConfig собирает конфигурацию по умолчанию, наполняя её константными таймаутами авторизации.
func NewDefaultAuthJWTConfig() *AuthJWTConfig {
	return NewAuthJWTConfig(
		DefaultAuthJWTEnabled,
		"",
		DefaultAuthJWTSigningMethod,
		"",
		DefaultAuthJWTLeeway,
		DefaultAuthJWTAccessTokenTTL,
		DefaultAuthJWTRefreshTokenTTL,
		"",
	)
}

// Validate осуществляет семантическую и системную проверку секретов и крипто-путей на этапе запуска (Bootstrap Phase).
// Защищает рантайм от запуска IAM контура с пустыми солями, нулевыми таймаутами или несмонтированными RSA-ключами.
func (jac *AuthJWTConfig) Validate() error {
	if !jac.Enabled {
		return nil
	}
	if jac.Secret == "" {
		return errs.NewConfigValidateError("jwt auth", "Secret", "is required", nil)
	}
	if _, err := jwt.GetSigningMethod(jac.SigningMethod); err != nil {
		return errs.NewConfigValidateError("jwt auth", "SigningMethod", "signing method validation failed", err)
	}
	if strings.TrimSpace(jac.Issuer) == "" {
		return errs.NewConfigValidateError("jwt auth", "Issuer", "is required", nil)
	}
	if jac.Leeway <= 0 {
		return errs.NewConfigValidateError("jwt auth", "Leeway", "must be greater than 0", nil)
	}
	if jac.AccessTokenTTL <= 0 {
		return errs.NewConfigValidateError("jwt auth", "AccessTokenTTL", "must be greater than 0", nil)
	}

	// Если путь к приватному ключу RSA задан — проактивно проверяем его физическое существование на диске (Fail-Fast)
	if strings.TrimSpace(jac.RSAPrivateKeyPath) != "" {
		if _, err := os.Stat(jac.RSAPrivateKeyPath); err != nil {
			if os.IsNotExist(err) {
				// Возвращаем типизированную ошибку валидации конфигурации фреймворка с привязкой исходного OS-сбоя
				return errs.NewConfigValidateError("jwt auth", "RSAPrivateKeyPath", "does not exist", err)
			}

			return errs.NewConfigValidateError("jwt auth", "RSAPrivateKeyPath", "path check failed", err)
		}
	}

	return nil
}

// BuildDefaultAuthJWTSettings auth settings builder
//
// # ВНИМАНИЕ!!!
//
// Сюда не входит:
//
// # TemplateKeyAuthJWTIssuer TemplateFlagAuthJWTIssuer
//
// TemplateKeyAuthJWTRSAPrivateKeyPath TemplateFlagAuthJWTRSAPrivateKeyPath
func BuildDefaultAuthJWTSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeyAuthJWTEnabled, keyPrefix), fmt.Sprintf(TemplateFlagAuthJWTEnabled, flagPrefix), DefaultAuthJWTEnabled, fmt.Sprintf("%s use jwt auth", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAuthJWTSecret, keyPrefix), fmt.Sprintf(TemplateFlagAuthJWTSecret, flagPrefix), DefaultAuthJWTSecret, fmt.Sprintf("%s secret", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAuthJWTSigningMethod, keyPrefix), fmt.Sprintf(TemplateFlagAuthJWTSigningMethod, flagPrefix), DefaultAuthJWTSigningMethod, fmt.Sprintf("%s signing method", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyAuthJWTIssuer, keyPrefix), fmt.Sprintf(TemplateFlagAuthJWTIssuer, flagPrefix), "", fmt.Sprintf("%s issuer", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAuthJWTLeeway, keyPrefix), fmt.Sprintf(TemplateFlagAuthJWTLeeway, flagPrefix), DefaultAuthJWTLeeway, fmt.Sprintf("%s leeway time interval", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAuthJWTAccessTokenTTL, keyPrefix), fmt.Sprintf(TemplateFlagAuthJWTAccessTokenTTL, flagPrefix), DefaultAuthJWTAccessTokenTTL, fmt.Sprintf("%s access token TTL", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyAuthJWTRefreshTokenTTL, keyPrefix), fmt.Sprintf(TemplateFlagAuthJWTRefreshTokenTTL, flagPrefix), DefaultAuthJWTRefreshTokenTTL, fmt.Sprintf("%s refresh token TTL", flagPrefix)),
		//NewSetting(fmt.Sprintf(TemplateKeyAuthJWTRSAPrivateKeyPath, keyPrefix), fmt.Sprintf(TemplateFlagAuthJWTRSAPrivateKeyPath, flagPrefix), "", fmt.Sprintf("%s RSA private key path", flagPrefix)),
	}
}
