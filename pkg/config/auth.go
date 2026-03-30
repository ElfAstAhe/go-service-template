package config

import (
	"os"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AuthConfig — секреты для JWT и безопасности
type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret" json:"jwt_secret,omitempty" yaml:"jwt_secret,omitempty"`
	JWTSigningMethod   string        `mapstructure:"jwt_signing_method" json:"jwt_signing_method,omitempty" yaml:"jwt_signing_method,omitempty"`
	AccessTokenTTL     time.Duration `mapstructure:"access_token_ttl" json:"access_token_ttl,omitempty" yaml:"access_token_ttl,omitempty"`
	RefreshTokenTTL    time.Duration `mapstructure:"refresh_token_ttl" json:"refresh_token_ttl,omitempty" yaml:"refresh_token_ttl,omitempty"`
	RSAPrivateKeyPath  string        `mapstructure:"rsa_private_key_path" json:"rsa_private_key_path,omitempty" yaml:"rsa_private_key_path,omitempty"`
	MasterPasswordSalt string        `mapstructure:"master_password_salt" json:"master_password_salt,omitempty" yaml:"master_password_salt,omitempty"`
}

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

func NewDefaultAuthConfig() *AuthConfig {
	return NewAuthConfig("", DefaultAuthSigningMethod, DefaultAuthAccessTokenTTL, DefaultAuthRefreshTokenTTL, "", "")
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
