package config

import (
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// flags templates
const (
	// TemplateFlagRunnerStopTimeout таймаут на остановку консьюмеров или серверов
	TemplateFlagRunnerStopTimeout string = "%s-stop-timeout"

	// TemplateFlagRunnerCloseTimeout таймаут на уничтожение внутренних сетевых ресурсов
	TemplateFlagRunnerCloseTimeout string = "%s-close-timeout"
)

// keys template
const (
	// TemplateKeyRunnerStopTimeout таймаут на остановку консьюмеров или серверов
	TemplateKeyRunnerStopTimeout string = "%s.stop_timeout"

	// TemplateKeyRunnerCloseTimeout таймаут на уничтожение внутренних сетевых ресурсов
	TemplateKeyRunnerCloseTimeout string = "%s.close_timeout"
)

// Временные лимиты по умолчанию для управления жизненным циклом асинхронных раннеров (Runners).
const (
	DefaultRunnerStopTimeout  time.Duration = 15 * time.Second // Дефолтный таймаут на остановку консьюмеров или серверов
	DefaultRunnerCloseTimeout time.Duration = 5 * time.Second  // Дефолтный таймаут на уничтожение внутренних сетевых ресурсов
)

// RunnerConfig инкапсулирует конфигурационные параметры временных лимитов (таймаутов)
// для управления фазами остановки и очистки ресурсов исполняемых компонентов (Runners).
type RunnerConfig struct {
	// StopTimeout Временной лимит на мягкую остановку приема трафика/сообщений
	StopTimeout time.Duration `mapstructure:"stop_timeout" json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty" toml:"stop_timeout,omitempty"`

	// CloseTimeout Временной лимит на деаллокацию и закрытие внутренних пулов/сокетов
	CloseTimeout time.Duration `mapstructure:"close_timeout" json:"close_timeout,omitempty" yaml:"close_timeout,omitempty" toml:"close_timeout,omitempty"`
}

// NewRunnerConfig — фабричный конструктор конфигурации раннеров.
func NewRunnerConfig(
	stopTimeout time.Duration,
	closeTimeout time.Duration,
) *RunnerConfig {
	return &RunnerConfig{
		StopTimeout:  stopTimeout,
		CloseTimeout: closeTimeout,
	}
}

// NewDefaultRunnerConfig собирает конфигурацию по умолчанию, наполняя её системными константными дефолтами.
func NewDefaultRunnerConfig() *RunnerConfig {
	return &RunnerConfig{
		StopTimeout:  DefaultRunnerStopTimeout,
		CloseTimeout: DefaultRunnerCloseTimeout,
	}
}

// Validate осуществляет семантическую проверку параметров раннера на этапе запуска микросервиса (Bootstrap Phase).
// Предотвращает запуск исполняемых компонентов с некорректными или отрицательными таймаутами жизненного цикла.
func (brc *RunnerConfig) Validate() error {
	if brc.StopTimeout < 0 {
		return errs.NewConfigValidateError("runner", "StopTimeout", "must be equal or greater zero", nil)
	}
	if brc.CloseTimeout < 0 {
		return errs.NewConfigValidateError("runner", "CloseTimeout", "must be equal or greater zero", nil)
	}

	return nil
}

// BuildDefaultRunnerSettings runner settings builder
func BuildDefaultRunnerSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeyRunnerStopTimeout, keyPrefix), fmt.Sprintf(TemplateFlagRunnerStopTimeout, flagPrefix), DefaultRunnerStopTimeout, fmt.Sprintf("%s stop timeout", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyRunnerCloseTimeout, keyPrefix), fmt.Sprintf(TemplateFlagRunnerCloseTimeout, flagPrefix), DefaultRunnerCloseTimeout, fmt.Sprintf("%s close timeout", flagPrefix)),
	}
}
