package config

import (
	"errors"
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
	App       *AppConfig            `mapstructure:"app" json:"app,omitempty" yaml:"app,omitempty"`
	Auth      *conf.AuthConfig      `mapstructure:"auth" json:"auth,omitempty" yaml:"auth,omitempty"`
	HTTP      *conf.HTTPConfig      `mapstructure:"http" json:"http,omitempty" yaml:"http,omitempty"`
	GRPC      *conf.GRPCConfig      `mapstructure:"grpc" json:"grpc,omitempty" yaml:"grpc,omitempty"`
	Log       *conf.LogConfig       `mapstructure:"log" json:"log,omitempty" yaml:"log,omitempty"`
	DB        *conf.DBConfig        `mapstructure:"db" json:"db,omitempty" yaml:"db,omitempty"`
	Telemetry *conf.TelemetryConfig `mapstructure:"telemetry" json:"telemetry,omitempty" yaml:"telemetry,omitempty"`
	//    Redis *RedisConfig `mapstructure:"redis"`
}

// linker params
var (
	AppName      string
	AppVersion   string
	AppBuildTime string
)

func NewConfig(
	app *AppConfig,
	auth *conf.AuthConfig,
	HTTP *conf.HTTPConfig,
	GRPC *conf.GRPCConfig,
	log *conf.LogConfig,
	db *conf.DBConfig,
	telemetry *conf.TelemetryConfig,
) *Config {
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

	var validateErrs []error
	for _, validator := range validators {
		validateErrs = append(validateErrs, validator.Validate())
	}
	err := errors.Join(validateErrs...)
	if err != nil {
		return errs.NewConfigValidateError("config", "config", "all config validation failed", err)
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

	// 3. Привязываем все флаги к Viper
	if err := bindFlags(pFlagSet, v); err != nil {
		return nil, err
	}

	// 4. Настройка Переменных окружения (ENV)
	// Используем твой механизм AutomaticEnv
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// 5. Поддержка ENV для пути к конфигу
	err = v.BindEnv(conf.FlagConfig, conf.EnvConfig)
	if err != nil {
		return nil, errs.NewConfigError("failed to bind env", err)
	}

	// 6. Чтение файла конфигурации
	cfgFile := v.GetString(conf.FlagConfig)
	v.SetConfigFile(cfgFile)

	if err := v.ReadInConfig(); err != nil {
		// Если файла нет — это предупреждение, но не фатальная ошибка (могут быть ENV)
		log.Printf("WARN: config file not found [%s]\n", cfgFile)
	}

	// 7. Маппинг в структуру
	var cfg = NewEmptyConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, errs.NewConfigError("failed to unmarshal config struct", err)
	}

	// 8. Итоговая валидация по всем слоям
	if err := cfg.Validate(); err != nil {
		return nil, errs.NewConfigError("config validation failed", err)
	}

	return cfg, nil
}

//goland:noinspection DuplicatedCode
func applyDefaults(v *viper.Viper) {
	// App
	v.SetDefault(conf.KeyAppEnv, conf.DefaultAppEnv)
	v.SetDefault(conf.KeyAppInitTimeout, conf.DefaultAppInitTimeout)
	v.SetDefault(conf.KeyAppStopTimeout, conf.DefaultAppStopTimeout)
	v.SetDefault(conf.KeyAppCloseTimeout, conf.DefaultAppCloseTimeout)

	// Auth
	v.SetDefault(conf.KeyAuthJWTSigningMethod, conf.DefaultAuthSigningMethod)
	v.SetDefault(conf.KeyAuthAccessTokenTTL, conf.DefaultAuthAccessTokenTTL)
	v.SetDefault(conf.KeyAuthRefreshTokenTTL, conf.DefaultAuthRefreshTokenTTL)

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
	v.SetDefault(conf.KeyDBDriver, conf.DefaultDBDriver)
	v.SetDefault(conf.KeyDBDSN, conf.DefaultDBDSN)
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
	res.String(conf.FlagConfig, "config/config.yaml", "path to config file")
	res.String(conf.FlagAppEnv, string(conf.DefaultAppEnv), "application environment")
	res.Duration(conf.FlagAppInitTimeout, conf.DefaultAppInitTimeout, "application init timeout")
	res.Duration(conf.FlagAppStopTimeout, conf.DefaultAppStopTimeout, "application stop timeout")
	res.Duration(conf.FlagAppCloseTimeout, conf.DefaultAppCloseTimeout, "application close timeout")

	// Auth
	res.String(conf.FlagAuthJWTSecret, "", "JWT secret")
	res.String(conf.FlagAuthJWTSigningMethod, conf.DefaultAuthSigningMethod, "JWT signing method")
	res.Duration(conf.FlagAuthAccessTokenTTL, conf.DefaultAuthAccessTokenTTL, "JWT access token TTL")
	res.Duration(conf.FlagAuthRefreshTokenTTL, conf.DefaultAuthRefreshTokenTTL, "JWT refresh token TTL")
	res.String(conf.FlagAuthRSAPrivateKeyPath, "", "RSA private key path")
	res.String(conf.FlagAuthMasterPasswordSalt, "", "master password salt")

	// HTTP
	res.String(conf.FlagHTTPAddress, conf.DefaultHTTPAddress, "http address")
	res.Duration(conf.FlagHTTPReadTimeout, conf.DefaultHTTPReadTimeout, "http read timeout")
	res.Duration(conf.FlagHTTPWriteTimeout, conf.DefaultHTTPWriteTimeout, "http write timeout")
	res.Duration(conf.FlagHTTPIdleTimeout, conf.DefaultHTTPIdleTimeout, "http idle timeout")
	res.Duration(conf.FlagHTTPShutdownTimeout, conf.DefaultHTTPShutdownTimeout, "http shutdown timeout")
	res.String(conf.FlagHTTPPrivateKeyPath, "", "http private key path")
	res.String(conf.FlagHTTPCertificatePath, "", "http certificate path")
	res.Bool(conf.FlagHTTPSecure, conf.DefaultHTTPSecure, "http secure mode")
	res.Int(conf.FlagHTTPMaxRequestBodySize, conf.DefaultHTTPMaxRequestBodySize, "http max request body size")

	// gRPC
	res.String(conf.FlagGRPCAddress, conf.DefaultGRPCAddress, "gRPC address")
	res.Duration(conf.FlagGRPCMaxConnIdle, conf.DefaultGRPCMaxConnIdle, "gRPC max connection idle timeout")
	res.Duration(conf.FlagGRPCMaxConnAge, conf.DefaultGRPCMaxConnAge, "gRPC max connection age timeout")
	res.Duration(conf.FlagGRPCMaxConnAgeGrace, conf.DefaultGRPCMaxConnAgeGrace, "gRPC max connection age grace timeout")
	res.Duration(conf.FlagGRPCTimeout, conf.DefaultGRPCTimeout, "gRPC timeout")
	res.Duration(conf.FlagGRPCKeepAliveTime, conf.DefaultGRPCKeepAliveTime, "gRPC keep alive timeout")
	res.Duration(conf.FlagGRPCKeepAliveTimeout, conf.DefaultGRPCKeepAliveTimeout, "gRPC keep alive timeout")
	res.Duration(conf.FlagGRPCShutdownTimeout, conf.DefaultGRPCShutdownTimeout, "gRPC shutdown timeout")

	// DB
	res.String(conf.FlagDBDSN, conf.DefaultDBDSN, "database dsn")
	res.String(conf.FlagDBDriver, conf.DefaultDBDriver, "database driver name/alias")
	res.Int(conf.FlagDBMaxOpenConns, conf.DefaultDBMaxOpenConns, "db max open connections")
	res.Int(conf.FlagDBMaxIdleConns, conf.DefaultDBMaxIdleConns, "db max idle connections")
	res.Duration(conf.FlagDBMaxIdleLifetime, conf.DefaultDBConnMaxIdleLifetime, "db max idle connection lifetime")
	res.Duration(conf.FlagDBConnTimeout, conf.DefaultDBConnTimeout, "db connection timeout)")

	// Log
	res.String(conf.FlagLogLevel, conf.DefaultLogLevel, "log level")
	res.String(conf.FlagLogFormat, conf.DefaultLogFormat, "log format")

	// Telemetry
	res.Bool(conf.FlagTelemetryEnabled, conf.DefaultTelemetryEnabled, "telemetry enabled")
	res.String(conf.FlagTelemetryServiceName, "", "telemetry service name")
	res.String(conf.FlagTelemetryExporterEndpoint, conf.DefaultTelemetryExporterEndpoint, "telemetry exporter endpoint")
	res.Float64(conf.FlagTelemetrySampleRate, conf.DefaultTelemetrySampleRate, "telemetry sample rate")
	res.Duration(conf.FlagTelemetryTimeout, conf.DefaultTelemetryTimeout, "telemetry timeout")

	// Добавь остальные pflag.String/Int/Duration для Redis, etc. ...
	// ..

	// Парсинг
	err = res.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	return res, err
}

func bindFlags(flags *pflag.FlagSet, v *viper.Viper) error {
	err := errors.Join(
		// App
		v.BindPFlag(conf.KeyAppEnv, flags.Lookup(conf.FlagAppEnv)),
		v.BindPFlag(conf.KeyAppInitTimeout, flags.Lookup(conf.FlagAppInitTimeout)),
		v.BindPFlag(conf.KeyAppStopTimeout, flags.Lookup(conf.FlagAppStopTimeout)),
		v.BindPFlag(conf.KeyAppCloseTimeout, flags.Lookup(conf.FlagAppCloseTimeout)),
		// Auth
		v.BindPFlag(conf.KeyAuthJWTSecret, flags.Lookup(conf.FlagAuthJWTSecret)),
		v.BindPFlag(conf.KeyAuthJWTSigningMethod, flags.Lookup(conf.FlagAuthJWTSigningMethod)),
		v.BindPFlag(conf.KeyAuthAccessTokenTTL, flags.Lookup(conf.FlagAuthAccessTokenTTL)),
		v.BindPFlag(conf.KeyAuthRefreshTokenTTL, flags.Lookup(conf.FlagAuthRefreshTokenTTL)),
		v.BindPFlag(conf.KeyAuthRSAPrivateKeyPath, flags.Lookup(conf.FlagAuthRSAPrivateKeyPath)),
		v.BindPFlag(conf.KeyAuthMasterPasswordSalt, flags.Lookup(conf.FlagAuthMasterPasswordSalt)),
		// HTTP
		v.BindPFlag(conf.KeyHTTPAddress, flags.Lookup(conf.FlagHTTPAddress)),
		v.BindPFlag(conf.KeyHTTPReadTimeout, flags.Lookup(conf.FlagHTTPReadTimeout)),
		v.BindPFlag(conf.KeyHTTPWriteTimeout, flags.Lookup(conf.FlagHTTPWriteTimeout)),
		v.BindPFlag(conf.KeyHTTPIdleTimeout, flags.Lookup(conf.FlagHTTPIdleTimeout)),
		v.BindPFlag(conf.KeyHTTPShutdownTimeout, flags.Lookup(conf.FlagHTTPShutdownTimeout)),
		v.BindPFlag(conf.KeyHTTPPrivateKeyPath, flags.Lookup(conf.FlagHTTPPrivateKeyPath)),
		v.BindPFlag(conf.KeyHTTPCertificatePath, flags.Lookup(conf.FlagHTTPCertificatePath)),
		v.BindPFlag(conf.KeyHTTPSecure, flags.Lookup(conf.FlagHTTPSecure)),
		v.BindPFlag(conf.KeyHTTPMaxRequestBodySize, flags.Lookup(conf.FlagHTTPMaxRequestBodySize)),
		// gRPC
		v.BindPFlag(conf.KeyGRPCAddress, flags.Lookup(conf.FlagGRPCAddress)),
		v.BindPFlag(conf.KeyGRPCMaxConnIdle, flags.Lookup(conf.FlagGRPCMaxConnIdle)),
		v.BindPFlag(conf.KeyGRPCMaxConnAge, flags.Lookup(conf.FlagGRPCMaxConnAge)),
		v.BindPFlag(conf.KeyGRPCMaxConnAgeGrace, flags.Lookup(conf.FlagGRPCMaxConnAgeGrace)),
		v.BindPFlag(conf.KeyGRPCTimeout, flags.Lookup(conf.FlagGRPCTimeout)),
		v.BindPFlag(conf.KeyGRPCKeepAliveTime, flags.Lookup(conf.FlagGRPCKeepAliveTime)),
		v.BindPFlag(conf.KeyGRPCKeepAliveTimeout, flags.Lookup(conf.FlagGRPCKeepAliveTimeout)),
		v.BindPFlag(conf.KeyGRPCShutdownTimeout, flags.Lookup(conf.FlagGRPCShutdownTimeout)),
		// Log
		v.BindPFlag(conf.KeyLogLevel, flags.Lookup(conf.FlagLogLevel)),
		v.BindPFlag(conf.KeyLogFormat, flags.Lookup(conf.FlagLogFormat)),
		// DB
		v.BindPFlag(conf.KeyDBDriver, flags.Lookup(conf.FlagDBDriver)),
		v.BindPFlag(conf.KeyDBDSN, flags.Lookup(conf.FlagDBDSN)),
		v.BindPFlag(conf.KeyDBMaxOpenConns, flags.Lookup(conf.FlagDBMaxOpenConns)),
		v.BindPFlag(conf.KeyDBMaxIdleConns, flags.Lookup(conf.FlagDBMaxIdleConns)),
		v.BindPFlag(conf.KeyDBConnMaxIdleLifetime, flags.Lookup(conf.FlagDBMaxIdleLifetime)),
		v.BindPFlag(conf.KeyDBConnTimeout, flags.Lookup(conf.FlagDBConnTimeout)),
		// Telemetry
		v.BindPFlag(conf.KeyTelemetryEnabled, flags.Lookup(conf.FlagTelemetryEnabled)),
		v.BindPFlag(conf.KeyTelemetryExporterEndpoint, flags.Lookup(conf.FlagTelemetryExporterEndpoint)),
		v.BindPFlag(conf.KeyTelemetryServiceName, flags.Lookup(conf.FlagTelemetryServiceName)),
		v.BindPFlag(conf.KeyTelemetrySampleRate, flags.Lookup(conf.FlagTelemetrySampleRate)),
		v.BindPFlag(conf.KeyTelemetryTimeout, flags.Lookup(conf.FlagTelemetryTimeout)),
	)
	if err != nil {
		return errs.NewConfigError("bind flags with keys", err)
	}

	return nil
}
