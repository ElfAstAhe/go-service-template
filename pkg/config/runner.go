package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type RunnerConfig struct {
	StopTimeout  time.Duration
	CloseTimeout time.Duration
}

func NewRunnerConfig(
	stopTimeout time.Duration,
	closeTimeout time.Duration,
) *RunnerConfig {
	return &RunnerConfig{
		StopTimeout:  stopTimeout,
		CloseTimeout: closeTimeout,
	}
}

func NewDefaultRunnerConfig() *RunnerConfig {
	return &RunnerConfig{
		StopTimeout:  DefaultRunnerStopTimeout,
		CloseTimeout: DefaultRunnerCloseTimeout,
	}
}

func (brc *RunnerConfig) Validate() error {
	if brc.StopTimeout < 0 {
		return errs.NewConfigValidateError("runner", "StopTimeout", "must be equal or greater zero", nil)
	}
	if brc.CloseTimeout < 0 {
		return errs.NewConfigValidateError("runner", "CloseTimeout", "must be equal or greater zero", nil)
	}

	return nil
}
