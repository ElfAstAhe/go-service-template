package config

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// RedisConfig — для кеша или очередей
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func NewRedisConfig(host, port, password string, db int) *RedisConfig {
	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}
}

func (rc *RedisConfig) Validate() error {
	if rc.Host == "" {
		return errs.NewConfigValidateError("redis", "host", "must not be empty", nil)
	}
	if rc.Port == "" {
		return errs.NewConfigValidateError("redis", "port", "must not be empty", nil)
	}
	if rc.Password == "" {
		return errs.NewConfigValidateError("redis", "password", "must not be empty", nil)
	}

	return nil
}
