package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// DBConfig — настройки основной реляционной базы данных
type DBConfig struct {
	Driver              string        `mapstructure:"driver"` // postgres, mysql, etc.
	DSN                 string        `mapstructure:"dsn"`
	MaxOpenConns        int           `mapstructure:"max_open_conns"`
	MaxIdleConns        int           `mapstructure:"max_idle_conns"`
	ConnMaxIdleLifetime time.Duration `mapstructure:"conn_max_idle_lifetime"`
	ConnTimeout         time.Duration `mapstructure:"conn_timeout"`
}

func NewDBConfig(driver, dsn string, maxOpenConns, maxIdleConns int, connMaxIdleLifetime, ConnTimeout time.Duration) *DBConfig {
	return &DBConfig{
		Driver:              driver,
		DSN:                 dsn,
		MaxOpenConns:        maxOpenConns,
		MaxIdleConns:        maxIdleConns,
		ConnMaxIdleLifetime: connMaxIdleLifetime,
		ConnTimeout:         ConnTimeout,
	}
}

func NewDefaultDBConfig() *DBConfig {
	return NewDBConfig("", "", defaultDBMaxOpenConns, defaultDBMaxIdleConns, defaultDBConnMaxLifetime, defaultDBConnTimeout)
}

func (dbc *DBConfig) Validate() error {
	if dbc.Driver == "" {
		return errs.NewConfigValidateError("db", "driver", "must not be empty", nil)
	}
	if dbc.DSN == "" {
		return errs.NewConfigValidateError("db", "dsn", "must not be empty", nil)
	}
	if dbc.MaxOpenConns <= 0 {
		return errs.NewConfigValidateError("db", "max_open_conns", "must be more than 0", nil)
	}
	if dbc.MaxIdleConns <= 0 {
		return errs.NewConfigValidateError("db", "max_idle_conns", "must be more than 0", nil)
	}
	if dbc.ConnMaxIdleLifetime <= 0 {
		return errs.NewConfigValidateError("db", "conn_max_idle_lifetime", "must be more than 0", nil)
	}
	if dbc.ConnTimeout <= 0 {
		return errs.NewConfigValidateError("db", "conn_timeout", "must be more than 0", nil)
	}

	return nil
}
