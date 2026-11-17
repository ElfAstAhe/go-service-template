package config

import (
	"os"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги подсистемы безопасности, шифрования и авторизации (IAM/Auth).
//
//nolint:gosec // G101 : это флаги (параметры)
const (
	FlagAuthJWTSecret          string = "auth-jwt-secret"           // CLI-флаг секретной соли HMAC подписи JWT токенов
	FlagAuthJWTSigningMethod   string = "auth-jwt-signing-method"   // CLI-флаг алгоритма подписи JWT (например, HS256)
	FlagAuthAccessTokenTTL     string = "auth-access-token-ttl"     // CLI-флаг времени жизни Access-токена
	FlagAuthRefreshTokenTTL    string = "auth-refresh-token-ttl"    // CLI-флаг времени жизни Refresh-токена
	FlagAuthRSAPrivateKeyPath  string = "auth-rsa-private-key-path" // CLI-флаг пути к приватному RSA-ключу на диске
	FlagAuthMasterPasswordSalt string = "auth-master-password-salt" // CLI-флаг соли для хэширования паролей пользователей
)

// ====================================================================
// Auth
// ====================================================================
// Строковые ключи Viper-маппинга для иерархического связывания секретов подсистемы безопасности и IAM.
//
//nolint:gosec // G101 : это ключи
const (
	KeyAuthJWTSecret          string = "auth.jwt_secret"
	KeyAuthJWTSigningMethod   string = "auth.jwt_signing_method"
	KeyAuthAccessTokenTTL     string = "auth.access_token_ttl"
	KeyAuthRefreshTokenTTL    string = "auth.refresh_token_ttl"
	KeyAuthRSAPrivateKeyPath  string = "auth.rsa_private_key_path"
	KeyAuthMasterPasswordSalt string = "auth.master_password_salt"
)

// ====================================================================
// access defaults
// ====================================================================
// Константные таймауты авторизации и параметры безопасности сессий по умолчанию.
const (
	DefaultAuthJWTSecret       string        = "123456"
	DefaultAuthSigningMethod   string        = "HS256"          // Метод подписи токенов по умолчанию: HMAC-SHA256
	DefaultAuthAccessTokenTTL  time.Duration = 15 * time.Minute // Время жизни короткоживущего токена доступа (Access Token)
	DefaultAuthRefreshTokenTTL time.Duration = 24 * time.Hour   // Время жизни сессионного токена обновления (Refresh Token)
)

// AuthConfig инкапсулирует криптографические секреты, алгоритмы подписи,
// временные лимиты сессий (JWT TTL) и параметры хэширования для подсистемы безопасности (IAM/Auth).
type AuthConfig struct {
	// JWTSecret секретный ключ (соль) для симметричной HMAC-подписи токенов
	JWTSecret string `mapstructure:"jwt_secret" json:"jwt_secret,omitempty" yaml:"jwt_secret,omitempty" toml:"jwt_secret,omitempty"`

	// JWTSigningMethod строковое имя алгоритма подписи (например, "HS256", "RS256")
	JWTSigningMethod string `mapstructure:"jwt_signing_method" json:"jwt_signing_method,omitempty" yaml:"jwt_signing_method,omitempty" toml:"jwt_signing_method,omitempty"`

	// AccessTokenTTL время жизни Access-токена (Default: 15m)
	AccessTokenTTL time.Duration `mapstructure:"access_token_ttl" json:"access_token_ttl,omitempty" yaml:"access_token_ttl,omitempty" toml:"access_token_ttl,omitempty"`

	// RefreshTokenTTL время жизни Refresh-токена (Default: 30d)
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" json:"refresh_token_ttl,omitempty" yaml:"refresh_token_ttl,omitempty" toml:"refresh_token_ttl,omitempty"`

	// RSAPrivateKeyPath путь в файловой системе к приватному RSA-ключу для асимметричной подписи токенов
	RSAPrivateKeyPath string `mapstructure:"rsa_private_key_path" json:"rsa_private_key_path,omitempty" yaml:"rsa_private_key_path,omitempty" toml:"rsa_private_key_path,omitempty"`

	// MasterPasswordSalt кастомная соль для хэширования паролей пользователей (Argon2/Bcrypt)
	MasterPasswordSalt string `mapstructure:"master_password_salt" json:"master_password_salt,omitempty" yaml:"master_password_salt,omitempty" toml:"master_password_salt,omitempty"`
}

// NewAuthConfig — фабричный конструктор конфигурации подсистемы безопасности.
func NewAuthConfig(
	JWTSecret string,
	JWTSigningMethod string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	RSAPrivateKeyPath string,
	MasterPasswordSalt string,
) *AuthConfig {
	return &AuthConfig{
		JWTSecret:          JWTSecret,
		JWTSigningMethod:   JWTSigningMethod,
		AccessTokenTTL:     accessTokenTTL,
		RefreshTokenTTL:    refreshTokenTTL,
		RSAPrivateKeyPath:  RSAPrivateKeyPath,
		MasterPasswordSalt: MasterPasswordSalt,
	}
}

// NewDefaultAuthConfig собирает конфигурацию по умолчанию, наполняя её константными таймаутами авторизации.
func NewDefaultAuthConfig() *AuthConfig {
	return NewAuthConfig("", DefaultAuthSigningMethod, DefaultAuthAccessTokenTTL, DefaultAuthRefreshTokenTTL, "", "")
}

// Validate осуществляет семантическую и системную проверку секретов и крипто-путей на этапе запуска (Bootstrap Phase).
// Защищает рантайм от запуска IAM контура с пустыми солями, нулевыми таймаутами или несмонтированными RSA-ключами.
func (ac *AuthConfig) Validate() error {
	if ac.JWTSecret == "" {
		return errs.NewConfigValidateError("auth", "JWTSecret", "must not be empty", nil)
	}
	if ac.AccessTokenTTL <= 0 {
		return errs.NewConfigValidateError("auth", "AccessTokenTTL", "must be greater than 0", nil)
	}

	// Если путь к приватному ключу RSA задан — проактивно проверяем его физическое существование на диске (Fail-Fast)
	if ac.RSAPrivateKeyPath != "" {
		if _, err := os.Stat(ac.RSAPrivateKeyPath); os.IsNotExist(err) {
			// Возвращаем типизированную ошибку валидации конфигурации фреймворка с привязкой исходного OS-сбоя
			return errs.NewConfigValidateError("auth", "RSAPrivateKeyPath", "does not exist", err)
		}
	}

	return nil
}

// BuildAuthSettings auth settings builder
func BuildAuthSettings() Settings {
	return Settings{
		NewSetting(KeyAuthJWTSecret, FlagAuthJWTSecret, DefaultAuthJWTSecret, "JWT secret"),
		NewSetting(KeyAuthJWTSigningMethod, FlagAuthJWTSigningMethod, DefaultAuthSigningMethod, "JWT signing method"),
		NewSetting(KeyAuthAccessTokenTTL, FlagAuthAccessTokenTTL, DefaultAuthAccessTokenTTL, "JWT access token TTL"),
		NewSetting(KeyAuthRefreshTokenTTL, FlagAuthRefreshTokenTTL, DefaultAuthRefreshTokenTTL, "JWT refresh token TTL"),
		NewSetting(KeyAuthRSAPrivateKeyPath, FlagAuthRSAPrivateKeyPath, "", "RSA private key path"),
		NewSetting(KeyAuthMasterPasswordSalt, FlagAuthMasterPasswordSalt, "", "master password salt"),
	}
}
