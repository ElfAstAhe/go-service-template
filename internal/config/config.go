package config

import (
	"errors"
	"log"
	"strings"

	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
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
		App: &AppConfig{
			AppConfig: &conf.AppConfig{},
		},
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
	var cfg = NewDefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, errs.NewConfigError("failed to unmarshal config struct", err)
	}

	// 8. Итоговая валидация по всем слоям
	if err := cfg.Validate(); err != nil {
		return nil, errs.NewConfigError("config validation failed", err)
	}

	return cfg, nil
}
