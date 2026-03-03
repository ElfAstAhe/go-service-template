package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config — корневой объект
type Config struct {
	App       *AppConfig            `mapstructure:"app"`
	Auth      *conf.AuthConfig      `mapstructure:"auth"`
	HTTP      *conf.HTTPConfig      `mapstructure:"http"`
	GRPC      *conf.GRPCConfig      `mapstructure:"grpc"`
	Log       *conf.LogConfig       `mapstructure:"log"`
	DB        *conf.DBConfig        `mapstructure:"db"` // <-- Универсальное имя
	Telemetry *conf.TelemetryConfig `mapstructure:"telemetry"`
	//    Redis *RedisConfig `mapstructure:"redis"`
}

// link params
var (
	AppName      string
	AppVersion   string
	AppBuildTime string
)

func NewConfig(app *AppConfig, auth *conf.AuthConfig, HTTP *conf.HTTPConfig, GRPC *conf.GRPCConfig, log *conf.LogConfig, db *conf.DBConfig, telemetry *conf.TelemetryConfig) *Config {
	return &Config{
		App:       app,
		Auth:      auth,
		HTTP:      HTTP,
		GRPC:      GRPC,
		Log:       log,
		DB:        db,
		Telemetry: telemetry,
	}
}

func NewDefaultConfig() *Config {
	return NewConfig(
		NewDefaultAppConfig(),
		conf.NewDefaultAuthConfig(),
		conf.NewDefaultHTTPConfig(),
		conf.NewDefaultGRPCConfig(),
		conf.NewDefaultLogConfig(),
		conf.NewDefaultDBConfig(),
		conf.NewDefaultTelemetryConfig(),
	)
}

func NewEmptyConfig() *Config {
	return &Config{
		App:       &AppConfig{},
		Auth:      &conf.AuthConfig{},
		HTTP:      &conf.HTTPConfig{},
		GRPC:      &conf.GRPCConfig{},
		Log:       &conf.LogConfig{},
		DB:        &conf.DBConfig{},
		Telemetry: &conf.TelemetryConfig{},
	}
}

func (c *Config) Validate() error {
	validators := []interface {
		Validate() error
	}{
		c.App,
		//		c.Auth,
		c.HTTP,
		c.GRPC,
		c.Log,
		c.DB,
		c.Telemetry,
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
		return nil, errs.NewConfigError("failed init flags", err)
	}

	// Привязываем все флаги к Viper
	if err := v.BindPFlags(pFlagSet); err != nil {
		return nil, errs.NewConfigError("failed to bind pFlags", err)
	}

	// 3. Настройка Переменных окружения (ENV)
	// Используем твой механизм AutomaticEnv
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 4. Поддержка ENV для пути к конфигу
	err = v.BindEnv(FlagConfig, EnvConfig)
	if err != nil {
		return nil, errs.NewConfigError("failed to bind env", err)
	}

	// 5. Чтение файла конфигурации
	cfgFile := v.GetString(FlagConfig)
	v.SetConfigFile(cfgFile)

	if err := v.ReadInConfig(); err != nil {
		// Если файла нет — это предупреждение, но не фатальная ошибка (могут быть ENV)
		log.Printf("WARN: config file not found [%s]\n", cfgFile)
	}

	// 6. Маппинг в структуру
	var cfg = NewEmptyConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, errs.NewConfigError("failed to unmarshal config struct", err)
	}

	// 7. Итоговая валидация по всем слоям
	if err := cfg.Validate(); err != nil {
		return nil, errs.NewConfigError("config validation failed", err)
	}

	return cfg, nil
}

func applyDefaults(v *viper.Viper) {
	// App
	v.SetDefault(keyAppEnv, defaultAppEnv)

	// HTTP
	v.SetDefault(conf.KeyHTTPAddress, conf.DefaultHTTPAddress)
	v.SetDefault(conf.KeyHTTPReadTimeout, conf.DefaultHTTPReadTimeout)
	v.SetDefault(conf.KeyHTTPWriteTimeout, conf.DefaultHTTPWriteTimeout)
	v.SetDefault(conf.KeyHTTPIdleTimeout, conf.DefaultHTTPIdleTimeout)
	v.SetDefault(conf.KeyHTTPShutdownTimeout, conf.DefaultHTTPShutdownTimeout)
	v.SetDefault(conf.KeyHTTPSecure, conf.DefaultHTTPSecure)
	v.SetDefault(conf.KeyHTTPMaxRequestBodySize, conf.DefaultHTTPMaxRequestBodySize)

	// gRPC
	v.SetDefault(conf.KeyGRPCAddress, conf.DefaultGRPCAddress)
	v.SetDefault(conf.KeyGRPCMaxConnIdle, conf.DefaultGRPCMaxConnIdle)
	v.SetDefault(conf.KeyGRPCMaxConnAge, conf.DefaultGRPCMaxConnAge)
	v.SetDefault(conf.KeyGRPCMaxConnAgeGrace, conf.DefaultGRPCMaxConnAgeGrace)
	v.SetDefault(conf.KeyGRPCTimeout, conf.DefaultGRPCTimeout)
	v.SetDefault(conf.KeyGRPCKeepAliveTime, conf.DefaultGRPCKeepAliveTime)
	v.SetDefault(conf.KeyGRPCKeepAliveTimeout, conf.DefaultGRPCKeepAliveTimeout)
	v.SetDefault(conf.KeyGRPCShutdownTimeout, conf.DefaultGRPCShutdownTimeout)

	// DB
	v.SetDefault(conf.KeyDBMaxOpenConns, conf.DefaultDBMaxOpenConns)
	v.SetDefault(conf.KeyDBMaxIdleConns, conf.DefaultDBMaxIdleConns)
	v.SetDefault(conf.KeyDBConnMaxIdleLifetime, conf.DefaultDBConnMaxIdleLifetime)
	v.SetDefault(conf.KeyDBConnTimeout, conf.DefaultDBConnTimeout)

	// Log
	v.SetDefault(conf.KeyLogLevel, conf.DefaultLogLevel)
	v.SetDefault(conf.KeyLogFormat, conf.DefaultLogFormat)

	// Telemetry
	v.SetDefault(conf.KeyTelemetryEnabled, conf.DefaultTelemetryEnabled)
	v.SetDefault(conf.KeyTelemetryExporterEndpoint, conf.DefaultTelemetryExporterEndpoint)
	v.SetDefault(conf.KeyTelemetrySampleRate, conf.DefaultTelemetrySampleRate)
	v.SetDefault(conf.KeyTelemetryTimeout, conf.DefaultTelemetryTimeout)

	// ... и так далее для всех критичных полей
}

func initFLags() (res *pflag.FlagSet, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Проверяем, является ли r ошибкой
			recoveryErr, ok := r.(error)
			if !ok {
				// Если это строка или что-то другое, приводим к виду error вручную
				recoveryErr = errs.NewConfigError(fmt.Sprintf("panic [%v] recovery", r), nil)
			}
			res = nil
			err = errs.NewConfigError("parse cli flags panic", recoveryErr)
		}
	}()

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
	res.String(FlagHTTPAddress, conf.DefaultHTTPAddress, "http address")
	res.Duration(FlagHTTPReadTimeout, conf.DefaultHTTPReadTimeout, "http read timeout")
	res.Duration(FlagHTTPWriteTimeout, conf.DefaultHTTPWriteTimeout, "http write timeout")
	res.Duration(FlagHTTPIdleTimeout, conf.DefaultHTTPIdleTimeout, "http idle timeout")
	res.Duration(FlagHTTPShutdownTimeout, conf.DefaultHTTPShutdownTimeout, "http shutdown timeout")
	res.String(FlagHTTPPrivateKeyPath, "", "http private key path")
	res.String(FlagHTTPCertificatePath, "", "http certificate path")
	res.Bool(FlagHTTPSecure, conf.DefaultHTTPSecure, "http secure mode")
	res.Int(FlagHTTPMaxRequestBodySize, conf.DefaultHTTPMaxRequestBodySize, "http max request body size")

	// gRPC
	res.String(FlagGRPCAddress, conf.DefaultGRPCAddress, "gRPC address")
	res.Duration(FlagGRPCMaxConnIdle, conf.DefaultGRPCMaxConnIdle, "gRPC max connection idle timeout")
	res.Duration(FlagGRPCMaxConnAge, conf.DefaultGRPCMaxConnAge, "gRPC max connection age timeout")
	res.Duration(FlagGRPCMaxConnAgeGrace, conf.DefaultGRPCMaxConnAgeGrace, "gRPC max connection age grace timeout")
	res.Duration(FlagGRPCTimeout, conf.DefaultGRPCTimeout, "gRPC timeout")
	res.Duration(FlagGRPCKeepAliveTime, conf.DefaultGRPCKeepAliveTime, "gRPC keep alive timeout")
	res.Duration(FlagGRPCKeepAliveTimeout, conf.DefaultGRPCKeepAliveTimeout, "gRPC keep alive timeout")
	res.Duration(FlagGRPCShutdownTimeout, conf.DefaultGRPCShutdownTimeout, "gRPC shutdown timeout")

	// DB
	res.String(FlagDBDSN, "", "database dsn")
	res.Int(FlagDBMaxOpenConns, conf.DefaultDBMaxOpenConns, "db max open connections")
	res.Int(FlagDBMaxIdleConns, conf.DefaultDBMaxIdleConns, "db max idle connections")
	res.Duration(FlagDBMaxIdleLifetime, conf.DefaultDBConnMaxIdleLifetime, "db max idle connection lifetime")
	res.Duration(FlagDBConnTimeout, conf.DefaultDBConnTimeout, "db connection timeout)")

	// Log
	res.String(FlagLogLevel, conf.DefaultLogLevel, "log level")
	res.String(FlagLogFormat, conf.DefaultLogFormat, "log format")

	// Telemetry
	res.Bool(FlagTelemetryEnabled, conf.DefaultTelemetryEnabled, "telemetry enabled")
	res.String(FlagTelemetryServiceName, "", "telemetry service name")
	res.String(FlagTelemetryExporterEndpoint, conf.DefaultTelemetryExporterEndpoint, "telemetry exporter endpoint")
	res.Float64(FlagTelemetrySampleRate, conf.DefaultTelemetrySampleRate, "telemetry sample rate")
	res.Duration(FlagTelemetryTimeout, conf.DefaultTelemetryTimeout, "telemetry timeout")

	// Добавь остальные pflag.String/Int/Duration для GRPC, Redis, etc. ...
	// ..

	// Парсинг
	err = res.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	return res, err
}
