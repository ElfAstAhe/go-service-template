package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги низкоуровневых параметров подключения и тюнинга пула реляционной СУБД.
const (
	FlagDBDriver          string = "db-driver"            // CLI-флаг типа SQL-драйвера (postgres, mysql, etc.)
	FlagDBDSN             string = "db-dsn"               // CLI-флаг строки подключения Data Source Name
	FlagDBMaxOpenConns    string = "db-max-open-conns"    // CLI-флаг жесткого верхнего лимита открытых соединений СУБД
	FlagDBMaxIdleConns    string = "db-max-idle-conns"    // CLI-флаг жесткого лимита удерживаемых idle-сокетов пула
	FlagDBMaxIdleLifetime string = "db-max-idle-lifetime" // CLI-флаг лимита времени жизни idle-сессий СУБД
	FlagDBConnTimeout     string = "db-conn-timeout"      // CLI-флаг таймаута установки первичной сетевой сессии
)

// ====================================================================
// DB
// ====================================================================
// Строковые ключи Viper-маппинга для иерархического связывания параметров реляционной СУБД.
const (
	KeyDBDriver          string = "db.driver"
	KeyDBDSN             string = "db.dsn"
	KeyDBMaxOpenConns    string = "db.max_open_conns"
	KeyDBMaxIdleConns    string = "db.max_idle_conns"
	KeyDBMaxIdleLifetime string = "db.conn_max_idle_lifetime"
	KeyDBConnTimeout     string = "db.conn_timeout"
)

// ====================================================================
// DB defaults (only pool settings)
// ====================================================================
// Конфигурация по умолчанию для пула соединений реляционной базы данных (Prod-Ready конфигурация).
const (
	DefaultDBDriver          string        = "pgx/v5"         // Инициализируется пустым, требуя явного переопределения (postgres/mysql)
	DefaultDBDSN             string        = ""               // Строка подключения (намеренно скрыта для безопасности)
	DefaultDBMaxOpenConns    int           = 32               // Верхний лимит одновременно открытых сетевых сокетов к СУБД
	DefaultDBMaxIdleConns    int           = 4                // Жесткий лимит удерживаемых простаивающих сокетов в пуле
	DefaultDBMaxIdleLifetime time.Duration = 60 * time.Second // Лимит жизненного цикла idle-сессий для защиты от закрытия балансировщиком
	DefaultDBConnTimeout     time.Duration = 30 * time.Second // Порог ожидания при первичной установке TCP-сессии с базой
)

// DBConfig — настройки основной реляционной базы данных
//
// Инкапсулирует конфигурационные параметры пула соединений, строковые dsn-секреты
// и сетевые лимиты для основной реляционной СУБД (PostgreSQL, MySQL).
type DBConfig struct {
	Driver              string        `mapstructure:"driver" json:"driver,omitempty" yaml:"driver,omitempty" toml:"driver,omitempty"`
	DSN                 string        `mapstructure:"dsn" json:"dsn,omitempty" yaml:"dsn,omitempty" toml:"dsn,omitempty"`
	MaxOpenConns        int           `mapstructure:"max_open_conns" json:"max_open_conns,omitempty" yaml:"max_open_conns,omitempty" toml:"max_open_conns,omitempty"`
	MaxIdleConns        int           `mapstructure:"max_idle_conns" json:"max_idle_conns,omitempty" yaml:"max_idle_conns,omitempty" toml:"max_idle_conns,omitempty"`
	ConnMaxIdleLifetime time.Duration `mapstructure:"conn_max_idle_lifetime" json:"conn_max_idle_lifetime,omitempty" yaml:"conn_max_idle_lifetime,omitempty" toml:"conn_max_idle_lifetime,omitempty"`
	ConnTimeout         time.Duration `mapstructure:"conn_timeout" json:"conn_timeout,omitempty" yaml:"conn_timeout,omitempty" toml:"conn_timeout,omitempty"`
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
		DefaultDBMaxIdleLifetime,
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

// BuildDBSettings DB settings builder
func BuildDBSettings() Settings {
	return Settings{
		NewSetting(KeyDBDriver, FlagDBDriver, DefaultDBDriver, "DB driver name (pgx, pgx/v5, postgres, mysql, etc."),
		NewSetting(KeyDBDSN, FlagDBDSN, DefaultDBDSN, "DB connection string"),
		NewSetting(KeyDBMaxOpenConns, FlagDBMaxOpenConns, DefaultDBMaxOpenConns, "DB max open connections"),
		NewSetting(KeyDBMaxIdleConns, FlagDBMaxIdleConns, DefaultDBMaxIdleConns, "DB max idle connections"),
		NewSetting(KeyDBMaxIdleLifetime, FlagDBMaxIdleLifetime, DefaultDBMaxIdleLifetime, "DB max idle connection lifetime"),
		NewSetting(KeyDBConnTimeout, FlagDBConnTimeout, DefaultDBConnTimeout, "DB connection timeout"),
	}
}
