package config

import (
	"os"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AuthConfig — секреты для JWT и безопасности
type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret"`
	AccessTokenTTL     time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL    time.Duration `mapstructure:"refresh_token_ttl"`
	RSAPrivateKeyPath  string        `mapstructure:"rsa_private_key_path"`
	MasterPasswordSalt string        `mapstructure:"master_password_salt"`
}

func NewAuthConfig(
	JWTSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	RSAPrivateKeyPath string,
	MasterPasswordSalt string,
) *AuthConfig {
	return &AuthConfig{
		JWTSecret:          JWTSecret,
		AccessTokenTTL:     accessTokenTTL,
		RefreshTokenTTL:    refreshTokenTTL,
		RSAPrivateKeyPath:  RSAPrivateKeyPath,
		MasterPasswordSalt: MasterPasswordSalt,
	}
}

func NewDefaultAuthConfig() *AuthConfig {
	return NewAuthConfig("", 0, 0, "", "")
}

func (ac *AuthConfig) Validate() error {
	if ac.JWTSecret == "" {
		return errs.NewConfigValidateError("auth", "JWTSecret", "must not be empty", nil)
	}
	if ac.AccessTokenTTL <= 0 {
		return errs.NewConfigValidateError("auth", "AccessTokenTTL", "must be greater than 0", nil)
	}
	if ac.RSAPrivateKeyPath != "" {
		if _, err := os.Stat(ac.RSAPrivateKeyPath); os.IsNotExist(err) {
			return errs.NewConfigValidateError("auth", "RSAPrivateKeyPath", "does not exist", err)
		}
	}

	return nil
}
