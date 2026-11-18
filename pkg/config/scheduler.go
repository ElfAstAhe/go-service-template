package config

import (
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// flags
const (
	// TemplateFlagSchedulerStartInterval первичная задержка (холодное смещение) перед самым первым тиком таймера
	TemplateFlagSchedulerStartInterval string = "%s-scheduler-start-interval"

	// TemplateFlagSchedulerScheduleInterval фиксированный интервал периодического повторения задач (период)
	TemplateFlagSchedulerScheduleInterval string = "%s-scheduler-schedule-interval"

	// TemplateFlagSchedulerStopTimeout Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	TemplateFlagSchedulerStopTimeout string = "%s-scheduler-stop-timeout"
)

// keys
const (
	// TemplateKeySchedulerStartInterval первичная задержка (холодное смещение) перед самым первым тиком таймера
	TemplateKeySchedulerStartInterval string = "%s.scheduler_start_interval"

	// TemplateKeySchedulerScheduleInterval фиксированный интервал периодического повторения задач (период)
	TemplateKeySchedulerScheduleInterval string = "%s.scheduler_schedule_interval"

	// TemplateKeySchedulerStopTimeout Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	TemplateKeySchedulerStopTimeout string = "%s.scheduler_stop_timeout"
)

// ====================================================================
// Scheduler Defaults
// ====================================================================
// Системные параметры по умолчанию для scheduler
const (
	// DefaultSchedulerStartInterval - первичная задержка (холодное смещение) перед самым первым тиком таймера
	DefaultSchedulerStartInterval time.Duration = time.Second
	// DefaultSchedulerScheduleInterval - фиксированный интервал периодического повторения задач (период)
	DefaultSchedulerScheduleInterval time.Duration = 5 * time.Second
	// DefaultSchedulerStopTimeout - Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	DefaultSchedulerStopTimeout time.Duration = 5 * time.Second
)

// SchedulerConfig инкапсулирует конфигурационные параметры интервалов запуска фоновых задач по расписанию.
type SchedulerConfig struct {
	// StartInterval Первичная задержка (холодное смещение) перед самым первым тиком таймера
	StartInterval time.Duration `mapstructure:"start_interval" json:"start_interval,omitempty" yaml:"start_interval,omitempty" toml:"start_interval,omitempty"`

	// ScheduleInterval Фиксированный интервал периодического повторения задач (период)
	ScheduleInterval time.Duration `mapstructure:"schedule_interval" json:"schedule_interval,omitempty" yaml:"schedule_interval,omitempty" toml:"schedule_interval,omitempty"`

	// StopTimeout Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	StopTimeout time.Duration `mapstructure:"stop_timeout" json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty" toml:"stop_timeout,omitempty"`
}

// NewSchedulerConfig — фабричный конструктор конфигурации планировщика.
func NewSchedulerConfig(
	startInterval time.Duration,
	scheduleInterval time.Duration,
	stopTimeout time.Duration,
) *SchedulerConfig {
	return &SchedulerConfig{
		StartInterval:    startInterval,
		ScheduleInterval: scheduleInterval,
		StopTimeout:      stopTimeout,
	}
}

// NewDefaultSchedulerConfig собирает конфигурацию worker pool по умолчанию с наполнением системными дефолтами.
func NewDefaultSchedulerConfig() *SchedulerConfig {
	return NewSchedulerConfig(
		DefaultSchedulerStartInterval,
		DefaultSchedulerScheduleInterval,
		DefaultSchedulerStopTimeout,
	)
}

// Validate выполняет строгую семантическую и математическую валидацию параметров scheduler на этапе запуска (Bootstrap Phase).
func (sc *SchedulerConfig) Validate() error {
	if sc.StartInterval <= 0 {
		return errs.NewConfigValidateError("scheduler", "start interval", "start interval is required", nil)
	}
	if sc.ScheduleInterval <= 0 {
		return errs.NewConfigValidateError("scheduler", "schedule interval", "schedule interval is required", nil)
	}
	if sc.StopTimeout <= 0 {
		return errs.NewConfigValidateError("scheduler", "stop timeout", "stop timeout is required", nil)
	}

	return nil
}

// BuildSchedulerSettings scheduler settings builder
func BuildSchedulerSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeySchedulerStartInterval, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerStartInterval, flagPrefix), DefaultSchedulerStartInterval, fmt.Sprintf("%s start interval", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeySchedulerScheduleInterval, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerScheduleInterval, flagPrefix), DefaultSchedulerScheduleInterval, fmt.Sprintf("%s schedule interval", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeySchedulerStopTimeout, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerStopTimeout, flagPrefix), DefaultSchedulerStopTimeout, fmt.Sprintf("%s stop timeout", flagPrefix)),
	}
}
