package config

import (
	"os"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AuthConfig инкапсулирует криптографические секреты, алгоритмы подписи,
// временные лимиты сессий (JWT TTL) и параметры хэширования для подсистемы безопасности (IAM/Auth).
type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret" json:"jwt_secret,omitempty" yaml:"jwt_secret,omitempty"`                               // Секретный ключ (соль) для симметричной HMAC-подписи токенов
	JWTSigningMethod   string        `mapstructure:"jwt_signing_method" json:"jwt_signing_method,omitempty" yaml:"jwt_signing_method,omitempty"`       // Строковое имя алгоритма подписи (например, "HS256", "RS256")
	AccessTokenTTL     time.Duration `mapstructure:"access_token_ttl" json:"access_token_ttl,omitempty" yaml:"access_token_ttl,omitempty"`             // Время жизни Access-токена (Default: 15m)
	RefreshTokenTTL    time.Duration `mapstructure:"refresh_token_ttl" json:"refresh_token_ttl,omitempty" yaml:"refresh_token_ttl,omitempty"`          // Время жизни Refresh-токена (Default: 30d)
	RSAPrivateKeyPath  string        `mapstructure:"rsa_private_key_path" json:"rsa_private_key_path,omitempty" yaml:"rsa_private_key_path,omitempty"` // Путь в файловой системе к приватному RSA-ключу для асимметричной подписи токенов
	MasterPasswordSalt string        `mapstructure:"master_password_salt" json:"master_password_salt,omitempty" yaml:"master_password_salt,omitempty"` // Кастомная соль для хэширования паролей пользователей (Argon2/Bcrypt)
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
