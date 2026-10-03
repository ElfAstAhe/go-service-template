package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// SchedulerConfig инкапсулирует конфигурационные параметры интервалов запуска фоновых задач по расписанию.
type SchedulerConfig struct {
	// StartInterval Первичная задержка (холодное смещение) перед самым первым тиком таймера
	StartInterval time.Duration `mapstructure:"start_interval" json:"start_interval,omitempty" yaml:"start_interval,omitempty"`

	// ScheduleInterval Фиксированный интервал периодического повторения задач (период)
	ScheduleInterval time.Duration `mapstructure:"schedule_interval" json:"schedule_interval,omitempty" yaml:"schedule_interval,omitempty"`

	// StopTimeout Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	StopTimeout time.Duration `mapstructure:"stop_timeout" json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty"`
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

// Validate выполняет строгую семантическую и математическую валидацию параметров worker pool на этапе запуска (Bootstrap Phase).
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
