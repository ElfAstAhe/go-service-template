package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги для переопределения базовых параметров жизненного цикла приложения.
const (
	FlagAppEnv          string = "env"               // CLI-флаг режима среды окружения (prod/dev/test)
	FlagAppInitTimeout  string = "app-init-timeout"  // CLI-флаг лимита времени на сборку IoC графа зависимостей
	FlagAppStopTimeout  string = "app-stop-timeout"  // CLI-флаг лимита времени на мягкую остановку серверов и воркеров
	FlagAppCloseTimeout string = "app-close-timeout" // CLI-флаг лимита времени на принудительную деаллокацию пулов СУБД
)

// Строковые ключи Viper-маппинга для иерархического связывания внутренних параметров ядра приложения.
const (
	KeyAppEnv          string = "app.env"
	KeyAppInitTimeout  string = "app.init_timeout"
	KeyAppStopTimeout  string = "app.stop_timeout"
	KeyAppCloseTimeout string = "app.close_timeout"
)

// Параметры по умолчанию для ядра жизненного цикла приложения.
const (
	DefaultAppEnv          AppEnv        = AppEnvDevelopment // Среда окружения по умолчанию (dev контур локальной разработки)
	DefaultAppInitTimeout  time.Duration = 30 * time.Second  // Дефолтный лимит на сборку и инициализацию IoC графа зависимостей
	DefaultAppStopTimeout  time.Duration = 20 * time.Second  // Дефолтный лимит на Graceful Shutdown серверов и воркеров
	DefaultAppCloseTimeout time.Duration = 5 * time.Second   // Дефолтный лимит на деаллокацию тяжелых пулов памяти и сокетов СУБД
)

// AppConfig инкапсулирует глобальные метаданные среды окружения и системные
// временные лимиты (таймауты) для управления фазами жизненного цикла микросервиса.
type AppConfig struct {
	// Env режим запуска рантайма (dev, prod, test)
	Env AppEnv `mapstructure:"env" json:"env,omitempty" yaml:"env,omitempty"`

	// InitTimeout временной потолок на сборку и инициализацию DI-контейнеров (Bootstrap)
	InitTimeout time.Duration `mapstructure:"init_timeout" json:"init_timeout,omitempty" yaml:"init_timeout,omitempty"`

	// StopTimeout временной потолок на мягкую остановку сетевых серверов и консьюмеров очередей
	StopTimeout time.Duration `mapstructure:"stop_timeout" json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty"`

	// CloseTimeout временной потолок на каскадное закрытие и деаллокацию пулов памяти/сокетов СУБД
	CloseTimeout time.Duration `mapstructure:"close_timeout" json:"close_timeout,omitempty" yaml:"close_timeout,omitempty"`
}

// NewAppConfig — фабричный конструктор центральной конфигурации приложения.
func NewAppConfig(
	env AppEnv,
	initTimeout time.Duration,
	stopTimeout time.Duration,
	closeTimeout time.Duration,
) *AppConfig {
	return &AppConfig{
		Env:          env,
		InitTimeout:  initTimeout,
		StopTimeout:  stopTimeout,
		CloseTimeout: closeTimeout,
	}
}

// NewDefaultAppConfig собирает базовую конфигурацию, наполняя её константными дефолтами фреймворка.
func NewDefaultAppConfig() *AppConfig {
	return NewAppConfig(
		DefaultAppEnv,
		DefaultAppInitTimeout,
		DefaultAppStopTimeout,
		DefaultAppCloseTimeout,
	)
}

// Validate осуществляет семантическую проверку системных параметров на этапе запуска (Bootstrap Phase).
// Предотвращает запуск микросервиса с невалидными типами окружений или отрицательными таймаутами жизненного цикла.
func (ac *AppConfig) Validate() error {
	if ac.Env == "" {
		return errs.NewConfigValidateError("app", "Env", "empty", nil)
	}
	// Валидируем соответствие переданной строки доступному пулу сред (dev/prod/test)
	if !ac.Env.Exists() {
		return errs.NewConfigValidateError("app", "env", "env value not match", nil)
	}
	if ac.InitTimeout < 0 {
		return errs.NewConfigValidateError("app", "InitTimeout", "must be equal or greater zero", nil)
	}
	if ac.StopTimeout < 0 {
		return errs.NewConfigValidateError("app", "StopTimeout", "must be equal or greater zero", nil)
	}
	if ac.CloseTimeout < 0 {
		return errs.NewConfigValidateError("app", "CloseTimeout", "must be equal or greater zero", nil)
	}

	return nil
}

// BuildAppSettings app settings builder
func BuildAppSettings() Settings {
	return Settings{
		NewSetting(KeyAppEnv, FlagAppEnv, DefaultAppEnv, "application environment (prod, dev, test)"),
		NewSetting(KeyAppInitTimeout, FlagAppInitTimeout, DefaultAppInitTimeout, "application init timeout"),
		NewSetting(KeyAppStopTimeout, FlagAppStopTimeout, DefaultAppStopTimeout, "application stop timeout"),
		NewSetting(KeyAppCloseTimeout, FlagAppCloseTimeout, DefaultAppCloseTimeout, "application close timeout"),
	}
}
