package config

import (
	"fmt"
	"time"

	pkgconf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	TemplateFlagTailCutterTailLimit    string = "%s-tail-limit"
	TemplateFlagTailCutterTailInterval string = "%s-tail-interval"
	TemplateFlagTailCutterTailCut      string = "%s-tail-cut"
)

const (
	TemplateKeyTailCutterTailLimit    string = "%s.tail_limit"
	TemplateKeyTailCutterTailInterval string = "%s.tail_interval"
	TemplateKeyTailCutterTailCut      string = "%s.tail_cut"
)

// TailCutterConfig default values
const (
	DefaultTailCutterStartInterval      time.Duration = 5 * time.Second
	DefaultTailCutterScheduleInterval   time.Duration = 1 * time.Minute
	DefaultTailCutterWorkerCount        int           = 2
	DefaultTailCutterDataCapacity       int           = 128
	DefaultTailCutterCompleteProcessing bool          = false
	DefaultTailCutterStopTimeout        time.Duration = 15 * time.Second
	DefaultTailCutterTailLimit          int           = 100
	DefaultTailCutterTailInterval       time.Duration = 182 * 24 * time.Hour // 182 days
	DefaultTailCutterTailCut            bool          = true
)

type TailCutterConfig struct {
	Dispatcher   *pkgconf.SchedulerDispatcherConfig `mapstructure:"dispatcher" json:"dispatcher,omitempty" yaml:"dispatcher,omitempty" toml:"dispatcher,omitempty"`
	TailLimit    int                                `mapstructure:"tail-limit" json:"tailLimit,omitempty" yaml:"tail-limit,omitempty" toml:"tail-limit,omitempty"`
	TailInterval time.Duration                      `mapstructure:"tail_interval" json:"tail_interval,omitempty" yaml:"tail_interval,omitempty" toml:"tail_interval,omitempty"`
	TailCut      bool                               `mapstructure:"tail_cut" json:"tail_cut,omitempty" yaml:"tail_cut,omitempty" toml:"tail_cut,omitempty"`
}

func NewTailCutterConfig(
	tailLimit int,
	tailInterval time.Duration,
	tailCut bool,
	dispatcher *pkgconf.SchedulerDispatcherConfig,
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
		pkgconf.NewSchedulerDispatcherConfig(
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

// BuildTailCutterSettings tail cutter settings builder
func BuildTailCutterSettings(keyPrefix, flagPrefix string) pkgconf.Settings {
	return append(
		pkgconf.BuildSchedulerDispatcherSettings(fmt.Sprintf("%s.dispatcher", keyPrefix), flagPrefix),
		pkgconf.NewSetting(fmt.Sprintf(TemplateKeyTailCutterTailLimit, keyPrefix), fmt.Sprintf(TemplateFlagTailCutterTailLimit, flagPrefix), DefaultTailCutterTailLimit, fmt.Sprintf("%s tail limit", flagPrefix)),
		pkgconf.NewSetting(fmt.Sprintf(TemplateKeyTailCutterTailInterval, keyPrefix), fmt.Sprintf(TemplateFlagTailCutterTailInterval, flagPrefix), DefaultTailCutterTailInterval, fmt.Sprintf("%s tail interval", flagPrefix)),
		pkgconf.NewSetting(fmt.Sprintf(TemplateKeyTailCutterTailCut, keyPrefix), fmt.Sprintf(TemplateFlagTailCutterTailCut, flagPrefix), DefaultTailCutterTailCut, fmt.Sprintf("%s tail cut", flagPrefix)),
	)
}
