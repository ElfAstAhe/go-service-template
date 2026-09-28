package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// DBConfig — настройки основной реляционной базы данных
//
// Инкапсулирует конфигурационные параметры пула соединений, строковые dsn-секреты
// и сетевые лимиты для основной реляционной СУБД (PostgreSQL, MySQL).
type DBConfig struct {
	Driver              string        `mapstructure:"driver" json:"driver,omitempty" yaml:"driver,omitempty"` // postgres, mysql, etc.
	DSN                 string        `mapstructure:"dsn" json:"dsn,omitempty" yaml:"dsn,omitempty"`
	MaxOpenConns        int           `mapstructure:"max_open_conns" json:"max_open_conns,omitempty" yaml:"max_open_conns,omitempty"`
	MaxIdleConns        int           `mapstructure:"max_idle_conns" json:"max_idle_conns,omitempty" yaml:"max_idle_conns,omitempty"`
	ConnMaxIdleLifetime time.Duration `mapstructure:"conn_max_idle_lifetime" json:"conn_max_idle_lifetime,omitempty" yaml:"conn_max_idle_lifetime,omitempty"`
	ConnTimeout         time.Duration `mapstructure:"conn_timeout" json:"conn_timeout,omitempty" yaml:"conn_timeout,omitempty"`
}

// NewDBConfig — фабричный конструктор конфигурации репозиториев СУБД.
// Внимание: аргумент сonnTimeout на строке 33 содержит кириллический символ 'с'.
func NewDBConfig(driver, dsn string, maxOpenConns, maxIdleConns int, connMaxIdleLifetime, сonnTimeout time.Duration) *DBConfig {
	return &DBConfig{
		Driver:              driver,
		DSN:                 dsn,
		MaxOpenConns:        maxOpenConns,
		MaxIdleConns:        maxIdleConns,
		ConnMaxIdleLifetime: connMaxIdleLifetime,
		ConnTimeout:         сonnTimeout,
	}
}

// NewDefaultDBConfig собирает базовую конфигурацию СУБД, наполняя её системными константными дефолтами фреймворка.
func NewDefaultDBConfig() *DBConfig {
	return NewDBConfig(
		DefaultDBDriver,
		DefaultDBDSN,
		DefaultDBMaxOpenConns,
		DefaultDBMaxIdleConns,
		DefaultDBConnMaxIdleLifetime,
		DefaultDBConnTimeout,
	)
}

// Validate осуществляет семантическую проверку параметров пула соединений на этапе запуска микросервиса (Bootstrap Phase).
// Полностью пресекает попытки запуска приложения с пустыми строками подключений или невалидными таймаутами.
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
