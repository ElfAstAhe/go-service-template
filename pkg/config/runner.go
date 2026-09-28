package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// RunnerConfig инкапсулирует конфигурационные параметры временных лимитов (таймаутов)
// для управления фазами остановки и очистки ресурсов исполняемых компонентов (Runners).
type RunnerConfig struct {
	// StopTimeout Временной лимит на мягкую остановку приема трафика/сообщений
	StopTimeout time.Duration
	// CloseTimeout Временной лимит на деаллокацию и закрытие внутренних пулов/сокетов
	CloseTimeout time.Duration
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
