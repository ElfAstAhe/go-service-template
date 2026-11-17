package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type TailCutterConfig struct {
	Dispatcher   *config.SchedulerDispatcherConfig `mapstructure:"dispatcher" json:"dispatcher,omitempty" yaml:"dispatcher,omitempty" toml:"dispatcher,omitempty"`
	TailLimit    int                               `mapstructure:"tail-limit" json:"tailLimit,omitempty" yaml:"tail-limit,omitempty" toml:"tail-limit,omitempty"`
	TailInterval time.Duration                     `mapstructure:"tail_interval" json:"tail_interval,omitempty" yaml:"tail_interval,omitempty" toml:"tail_interval,omitempty"`
	TailCut      bool                              `mapstructure:"tail_cut" json:"tail_cut,omitempty" yaml:"tail_cut,omitempty" toml:"tail_cut,omitempty"`
}

func NewTailCutterConfig(
	tailLimit int,
	tailInterval time.Duration,
	tailCut bool,
	dispatcher *config.SchedulerDispatcherConfig,
) *TailCutterConfig {
	return &TailCutterConfig{
		Dispatcher:   dispatcher,
		TailInterval: tailInterval,
		TailCut:      tailCut,
	}
}

func NewDefaultTailCutterConfig() *TailCutterConfig {
	return NewTailCutterConfig(
		DefaultTailCutterTailLimit,
		DefaultTailCutterTailInterval,
		DefaultTailCutterTailCut,
		config.NewSchedulerDispatcherConfig(
			DefaultTailCutterWorkerCount,
			DefaultTailCutterDataCapacity,
			DefaultTailCutterCompleteProcessing,
			DefaultTailCutterStartInterval,
			DefaultTailCutterScheduleInterval,
			DefaultTailCutterStopTimeout,
		),
	)
}

// Validate метод проверки настроечных значений
func (tcc *TailCutterConfig) Validate() error {
	if err := tcc.Dispatcher.Validate(); err != nil {
		return errs.NewConfigValidateError("tail_cutter", "dispatcher", "validation failed", err)
	}
	if tcc.TailLimit <= 0 {
		return errs.NewConfigValidateError("tail_cutter", "tail limit", "must be greater than zero", nil)
	}
	if tcc.TailInterval <= 0 {
		return errs.NewConfigValidateError("tail_cutter", "tail interval", "must be greater zero", nil)
	}

	return nil
}
