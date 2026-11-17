package config

import (
	"errors"
	"log"
	"strings"

	pkgconf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config — корневой объект
type Config struct {
	App       *AppConfig               `mapstructure:"app" json:"app,omitempty" yaml:"app,omitempty" toml:"app,omitempty"`
	Auth      *pkgconf.AuthConfig      `mapstructure:"auth" json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	HTTP      *pkgconf.HTTPConfig      `mapstructure:"http" json:"http,omitempty" yaml:"http,omitempty" toml:"http,omitempty"`
	GRPC      *pkgconf.GRPCConfig      `mapstructure:"grpc" json:"grpc,omitempty" yaml:"grpc,omitempty" toml:"grpc,omitempty"`
	Log       *pkgconf.LogConfig       `mapstructure:"log" json:"log,omitempty" yaml:"log,omitempty" toml:"log,omitempty"`
	DB        *pkgconf.DBConfig        `mapstructure:"db" json:"db,omitempty" yaml:"db,omitempty" toml:"db,omitempty"`
	Telemetry *pkgconf.TelemetryConfig `mapstructure:"telemetry" json:"telemetry,omitempty" yaml:"telemetry,omitempty" toml:"telemetry,omitempty"`
}

var _ pkgconf.Validator = (*Config)(nil)

// linker params
var (
	AppName      string
	AppVersion   string
	AppBuildTime string
)

func NewConfig(
	app *AppConfig,
	auth *pkgconf.AuthConfig,
	HTTP *pkgconf.HTTPConfig,
	GRPC *pkgconf.GRPCConfig,
	log *pkgconf.LogConfig,
	db *pkgconf.DBConfig,
	telemetry *pkgconf.TelemetryConfig,
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
		pkgconf.NewDefaultAuthConfig(),
		pkgconf.NewDefaultHTTPConfig(),
		pkgconf.NewDefaultGRPCConfig(),
		pkgconf.NewDefaultLogConfig(),
		pkgconf.NewDefaultDBConfig(),
		pkgconf.NewDefaultTelemetryConfig(),
	)
}

func (c *Config) Validate() error {
	validators := []pkgconf.Validator{
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
	// 1. Instances.
	settings := buildSettings()
	v := viper.New()
	fs := pflag.NewFlagSet("app-cmd-flags", pflag.PanicOnError)

	// 2. Init flags.
	if err := settings.InitFlags(fs); err != nil {
		return nil, errs.NewConfigError("flags initialization failed", err)
	}

	// 3. Init keys.
	if err := settings.InitKeys(v); err != nil {
		return nil, errs.NewConfigError("keys initialization failed", err)
	}

	// 4. Binds.
	if err := settings.BindKeysAndFlags(v, fs); err != nil {
		return nil, errs.NewConfigError("keys and flags binding failed", err)
	}

	// 5. Настройка Переменных окружения (ENV). Используем твой механизм AutomaticEnv
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// 6. Поддержка ENV для пути к конфигу
	if err := v.BindEnv(pkgconf.FlagConfig, pkgconf.EnvConfig); err != nil {
		return nil, errs.NewConfigError("failed to bind env", err)
	}

	// 7. Чтение файла конфигурации
	confPath := v.GetString(pkgconf.FlagConfig)
	v.SetConfigFile(confPath)

	if err := v.ReadInConfig(); err != nil {
		// Если файла нет — это предупреждение, но не фатальная ошибка (могут быть ENV)
		log.Printf("WARN: config file not found [%s]\n", confPath)
	}

	// 8. Маппинг в структуру
	var res = NewDefaultConfig()
	if err := v.Unmarshal(res); err != nil {
		return nil, errs.NewConfigError("failed to unmarshal config struct", err)
	}

	// 9. Итоговая валидация по всем слоям
	if err := res.Validate(); err != nil {
		return nil, errs.NewConfigError("config validation failed", err)
	}

	return res, nil
}

func buildSettings() pkgconf.Settings {
	var builders = []func() pkgconf.Settings{
		BuildAppSettings,
		pkgconf.BuildAuthSettings,
		pkgconf.BuildHTTPSettings,
		pkgconf.BuildGRPCSettings,
		pkgconf.BuildLogSettings,
		pkgconf.BuildDBSettings,
		pkgconf.BuildTelemetrySettings,
	}
	var res pkgconf.Settings
	for _, builder := range builders {
		res = append(res, builder()...)
	}

	return res
}
