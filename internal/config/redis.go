package config

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// RedisConfig — для кеша или очередей
//
// Инкапсулирует параметры подключения к распределенному in-memory хранилищу данных Redis.
// Используется подсистемами кэширования L2-репозиториев и распределенных блокировок.
type RedisConfig struct {
	Host     string `mapstructure:"host"`     // Сетевой адрес хоста или IP-адрес инстанса Redis
	Port     string `mapstructure:"port"`     // Сетевой порт подключения (например, "6379")
	Password string `mapstructure:"password"` // Пароль авторизации (инжектируется из секретов инфраструктуры)
	DB       int    `mapstructure:"db"`       // Индекс целевой логической базы данных Redis (например, 0)
}

// NewRedisConfig — фабричный конструктор конфигурации подключения к Redis.
func NewRedisConfig(host, port, password string, db int) *RedisConfig {
	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}
}

// Validate осуществляет семантическую проверку конфигурационных параметров на этапе стартапа (Bootstrap Phase).
// Предотвращает запуск кэш-слоя с пустыми адресами или незаданными паролями авторизации.
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
