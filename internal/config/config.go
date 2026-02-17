package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config — корневой объект
type Config struct {
	App  *AppConfig  `mapstructure:"app"`
	Auth *AuthConfig `mapstructure:"auth"`
	HTTP *HTTPConfig `mapstructure:"http"`
	//    GRPC  *GRPCConfig  `mapstructure:"grpc"`
	Log *LogConfig `mapstructure:"log"`
	DB  *DBConfig  `mapstructure:"db"` // <-- Универсальное имя
	//    Redis *RedisConfig `mapstructure:"redis"`
}

// link params
var (
	AppName      string
	AppVersion   string
	AppBuildTime string
)

func NewConfig(app *AppConfig, auth *AuthConfig, HTTP *HTTPConfig, log *LogConfig, db *DBConfig) *Config {
	return &Config{
		App:  app,
		Auth: auth,
		HTTP: HTTP,
		Log:  log,
		DB:   db,
	}
}

func NewDefaultConfig() *Config {
	return NewConfig(
		NewDefaultAppConfig(),
		NewDefaultAuthConfig(),
		NewDefaultHTTPConfig(),
		NewDefaultLogConfig(),
		NewDefaultDBConfig(),
	)
}

func (c *Config) Validate() error {
	validators := []interface {
		Validate() error
	}{
		c.App, c.Auth, c.HTTP, c.Log, c.DB,
	}

	for _, validator := range validators {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Load собирает конфигурацию из: Flags -> ENV -> YAML -> Defaults
func Load() (*Config, error) {
	v := viper.New()

	// 1. Установка значений по умолчанию (Defaults)
	applyDefaults(v)

	// 2. Настройка Флагов (pflag для длинных имен --port)
	pFlagSet, err := initFLags()
	if err != nil {
		return nil, errs.NewConfigError("error init flags", err)
	}

	// Привязываем все флаги к Viper
	if err := v.BindPFlags(pFlagSet); err != nil {
		return nil, errs.NewConfigError("failed to bind pflags", err)
	}

	// 2. Настройка Переменных окружения (ENV)
	// Используем твой механизм AutomaticEnv
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 3. Поддержка ENV для пути к конфигу
	err = v.BindEnv(FlagConfig, EnvConfig)
	if err != nil {
		return nil, errs.NewConfigError("failed to bind env", err)
	}

	// 5. Чтение файла конфигурации
	cfgFile := v.GetString(FlagConfig)
	v.SetConfigFile(cfgFile)

	if err := v.ReadInConfig(); err != nil {
		// Если файла нет — это предупреждение, но не фатальная ошибка (могут быть ENV)
		fmt.Printf("Warning: config file not found: %s\n", cfgFile)
	}

	// 6. Маппинг в структуру
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errs.NewConfigError("failed to unmarshal config struct", err)
	}

	// 7. Итоговая валидация по всем слоям
	if err := cfg.Validate(); err != nil {
		return nil, errs.NewConfigError("config validation failed", err)
	}

	return &cfg, nil
}

func applyDefaults(v *viper.Viper) {
	// App
	v.SetDefault(keyAppEnv, defaultAppEnv)

	// HTTP
	v.SetDefault(keyHTTPAddress, defaultHTTPAddress)
	v.SetDefault(keyHTTPReadTimeout, defaultHTTPReadTimeout)
	v.SetDefault(keyHTTPWriteTimeout, defaultHTTPWriteTimeout)
	v.SetDefault(keyHTTPIdleTimeout, defaultHTTPIdleTimeout)
	v.SetDefault(keyHTTPShutdownTimeout, defaultHTTPShutdownTimeout)
	v.SetDefault(keyHTTPSecure, defaultHTTPSecure)

	// gRPC
	v.SetDefault(keyGRPCAddress, defaultGRPCAddress)
	v.SetDefault(keyGRPCTimeout, defaultGRPCTimeout)

	// DB
	v.SetDefault(keyDBMaxOpenConns, defaultDBMaxOpenConns)
	v.SetDefault(keyDBMaxIdleConns, defaultDBMaxIdleConns)
	v.SetDefault(keyDBConnMaxIdleLifetime, defaultDBConnMaxIdleLifetime)
	v.SetDefault(keyDBConnTimeout, defaultDBConnTimeout)

	// Log
	v.SetDefault(keyLogLevel, defaultLogLevel)
	v.SetDefault(keyLogFormat, defaultLogFormat)

	// ... и так далее для всех критичных полей
}

func initFLags() (res *pflag.FlagSet, err error) {
	res = pflag.NewFlagSet("cmd flags", pflag.PanicOnError)

	// Используем константы Flag...
	res.String(FlagConfig, "config/config.yaml", "path to config file")
	res.String(FlagAppEnv, string(defaultAppEnv), "application environment")

	// Auth
	res.String(FlagAuthJWTSecret, "", "JWT secret")
	res.Duration(FlagAuthAccessTokenTTL, 0, "JWT token TTL")
	res.Duration(FlagAuthRefreshTokenTTL, 0, "JWT refresh TTL")
	res.String(FlagAuthRSAPrivateKeyPath, "", "RSA private key path")
	res.String(FlagAuthMasterPasswordSalt, "", "master password salt")

	// HTTP
	res.String(FlagHTTPAddress, defaultHTTPAddress, "http address")
	res.Duration(FlagHTTPReadTimeout, defaultHTTPReadTimeout, "http read timeout")
	res.Duration(FlagHTTPWriteTimeout, defaultHTTPWriteTimeout, "http write timeout")
	res.Duration(FlagHTTPIdleTimeout, defaultHTTPIdleTimeout, "http idle timeout")
	res.Duration(FlagHTTPShutdownTimeout, defaultHTTPShutdownTimeout, "http shutdown timeout")
	res.String(FlagHTTPPrivateKeyPath, "", "http private key path")
	res.String(FlagHTTPCertificatePath, "", "http certificate path")
	res.Bool(FlagHTTPSecure, defaultHTTPSecure, "http secure mode")

	// gRPC
	res.String(FlagGRPCAddress, defaultGRPCAddress, "gRPC address")
	res.Duration(FlagGRPCMaxConnIdle, 0, "gRPC max connection idle timeout")
	res.Duration(FlagGRPCMaxConnAge, 0, "gRPC max connection age timeout")
	res.Duration(FlagGRPCTimeout, defaultGRPCTimeout, "gRPC timeout")
	res.Duration(FlagGRPCKeepAliveTime, 0, "gRPC keep alive timeout")
	res.Duration(FlagGRPCKeepAliveTimeout, 0, "gRPC keep alive timeout")

	// DB
	res.String(FlagDBDSN, "", "database dsn")
	res.Int(FlagDBMaxOpenConns, defaultDBMaxOpenConns, "db max open connections")
	res.Int(FlagDBMaxIdleConns, defaultDBMaxIdleConns, "db max idle connections")
	res.Duration(FlagDBMaxIdleLifetime, defaultDBConnMaxIdleLifetime, "db max idle connection lifetime")
	res.Duration(FlagDBConnTimeout, defaultDBConnTimeout, "db connection timeout")

	// Log
	res.String(FlagLogLevel, defaultLogLevel, "log level")
	res.String(FlagLogFormat, defaultLogFormat, "log format")

	// Добавь остальные pflag.String/Int/Duration для GRPC, Redis, etc. ...
	// ..

	// Парсинг
	err = res.Parse(os.Args[1:])
	defer func() {
		if r := recover(); r != nil {
			// Проверяем, является ли r ошибкой
			recoveryErr, ok := r.(error)
			if !ok {
				// Если это строка или что-то другое, приводим к виду error вручную
				recoveryErr = fmt.Errorf("%v", r)
			}
			res = nil
			err = errs.NewConfigError("parse cli flags panic", recoveryErr)
		}
	}()

	return res, err
}
