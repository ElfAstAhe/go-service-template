package config

import (
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// flags
const (
	TemplateFlagSchedulerDispatcherPoolWorkerCount           string = "%s-worker-count"
	TemplateFlagSchedulerDispatcherPoolDataCapacity          string = "%s-data-capacity"
	TemplateFlagSchedulerDispatcherCompleteProcess           string = "%s-complete-process"
	TemplateFlagSchedulerDispatcherSchedulerStartInterval    string = "%s-start-interval"
	TemplateFlagSchedulerDispatcherSchedulerScheduleInterval string = "%s-schedule-interval"
	TemplateFlagSchedulerDispatcherSchedulerStopTimeout      string = "%s-stop-timeout"
)

// keys
const (
	TemplateKeySchedulerDispatcherPoolWorkerCount           string = "%s.worker_count"
	TemplateKeySchedulerDispatcherPoolDataCapacity          string = "%s.data_capacity"
	TemplateKeySchedulerDispatcherPoolCompleteProcess       string = "%s.complete_process"
	TemplateKeySchedulerDispatcherSchedulerStartInterval    string = "%s.start_interval"
	TemplateKeySchedulerDispatcherSchedulerScheduleInterval string = "%s.schedule_interval"
	TemplateKeySchedulerDispatcherSchedulerStopTimeout      string = "%s.stop_timeout"
)

// ====================================================================
// Scheduler Dispatcher Defaults
// ====================================================================
// Системные параметры по умолчанию для scheduler
const (
	// DefaultSchedulerDispatcherPoolWorkerCount - кол-во обработчиков
	DefaultSchedulerDispatcherPoolWorkerCount int = 2
	// DefaultSchedulerDispatcherPoolDataCapacity - емкость внутреннего фонового буфера
	DefaultSchedulerDispatcherPoolDataCapacity int = 100
	// DefaultSchedulerDispatcherPoolCompleteProcess - признак завершения обработки при закрытии
	DefaultSchedulerDispatcherPoolCompleteProcess bool = true
	// DefaultSchedulerDispatcherSchedulerStartInterval - первичная задержка (холодное смещение) перед самым первым тиком таймера
	DefaultSchedulerDispatcherSchedulerStartInterval time.Duration = time.Second
	// DefaultSchedulerDispatcherSchedulerScheduleInterval - фиксированный интервал периодического повторения задач (период)
	DefaultSchedulerDispatcherSchedulerScheduleInterval time.Duration = 5 * time.Second
	// DefaultSchedulerDispatcherStopTimeout - Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	DefaultSchedulerDispatcherStopTimeout time.Duration = 5 * time.Second
)

// SchedulerDispatcherConfig инкапсулирует конфигурационные параметры емкости и политик останова пула воркеров.
type SchedulerDispatcherConfig struct {
	// WorkerCount Количество параллельно запущенных горутин-обработчиков
	WorkerCount int `mapstructure:"worker_count" json:"worker_count,omitempty" yaml:"worker_count,omitempty" toml:"worker_count,omitempty"`

	// DataCapacity Буферная емкость внутреннего канала задач (Backpressure window)
	DataCapacity int `mapstructure:"data_capacity" json:"data_capacity,omitempty" yaml:"data_capacity,omitempty" toml:"data_capacity,omitempty"`

	// CompleteProcess Флаг: вычитывать ли буфер до конца при закрытии канала (true) или тушить экстренно (false)
	CompleteProcess bool `mapstructure:"complete_process" json:"complete_process,omitempty" yaml:"complete_process,omitempty" toml:"complete_process,omitempty"`

	// StartInterval Первичная задержка (холодное смещение) перед самым первым тиком таймера
	StartInterval time.Duration `mapstructure:"start_interval" json:"start_interval,omitempty" yaml:"start_interval,omitempty" toml:"start_interval,omitempty"`

	// ScheduleInterval Фиксированный интервал периодического повторения задач (период)
	ScheduleInterval time.Duration `mapstructure:"schedule_interval" json:"schedule_interval,omitempty" yaml:"schedule_interval,omitempty" toml:"schedule_interval,omitempty"`

	// StopTimeout Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	StopTimeout time.Duration `mapstructure:"stop_timeout" json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty" toml:"stop_timeout,omitempty"`
}

// NewSchedulerDispatcherConfig acts as a full generic factory constructor allocating composed scheduler dispatcher attributes.
func NewSchedulerDispatcherConfig(
	workerCount int,
	dataCapacity int,
	completeProcess bool,
	startInterval time.Duration,
	scheduleInterval time.Duration,
	stopTimeout time.Duration,
) *SchedulerDispatcherConfig {
	return &SchedulerDispatcherConfig{
		WorkerCount:      workerCount,
		DataCapacity:     dataCapacity,
		CompleteProcess:  completeProcess,
		StartInterval:    startInterval,
		ScheduleInterval: scheduleInterval,
		StopTimeout:      stopTimeout,
	}
}

// NewDefaultSchedulerDispatcherConfig compiles a new fallback configuration layout populated with explicit framework defaults keys indices.
func NewDefaultSchedulerDispatcherConfig() *SchedulerDispatcherConfig {
	return NewSchedulerDispatcherConfig(
		DefaultSchedulerDispatcherPoolWorkerCount,
		DefaultSchedulerDispatcherPoolDataCapacity,
		DefaultSchedulerDispatcherPoolCompleteProcess,
		DefaultSchedulerDispatcherSchedulerStartInterval,
		DefaultSchedulerDispatcherSchedulerScheduleInterval,
		DefaultSchedulerDispatcherStopTimeout,
	)
}

// Validate executes mathematical threshold bounds validations prior to launching async ticker routing infrastructure.
func (sdc *SchedulerDispatcherConfig) Validate() error {
	if sdc.WorkerCount <= 0 {
		return errs.NewConfigValidateError("scheduler dispatcher", "worker count", "must be greater than 0", nil)
	}
	if sdc.DataCapacity <= 0 {
		return errs.NewConfigValidateError("scheduler dispatcher", "data capacity", "must be greater than 0", nil)
	}
	if sdc.StartInterval <= 0 {
		return errs.NewConfigValidateError("scheduler dispatcher", "start interval", "start interval is required", nil)
	}
	if sdc.ScheduleInterval <= 0 {
		return errs.NewConfigValidateError("scheduler dispatcher", "schedule interval", "schedule interval is required", nil)
	}
	if sdc.StopTimeout <= 0 {
		return errs.NewConfigValidateError("scheduler dispatcher", "stop timeout", "stop timeout is required", nil)
	}

	return nil
}

// BuildDefaultSchedulerDispatcherSettings scheduler dispatcher settings builder
func BuildDefaultSchedulerDispatcherSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeySchedulerDispatcherPoolWorkerCount, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerDispatcherPoolWorkerCount, flagPrefix), DefaultSchedulerDispatcherPoolWorkerCount, fmt.Sprintf("%s worker count", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeySchedulerDispatcherPoolDataCapacity, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerDispatcherPoolDataCapacity, flagPrefix), DefaultSchedulerDispatcherPoolDataCapacity, fmt.Sprintf("%s data capacity", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeySchedulerDispatcherPoolCompleteProcess, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerDispatcherCompleteProcess, flagPrefix), DefaultSchedulerDispatcherPoolCompleteProcess, fmt.Sprintf("%s complete process", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeySchedulerDispatcherSchedulerStartInterval, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerDispatcherSchedulerStartInterval, flagPrefix), DefaultSchedulerDispatcherSchedulerStartInterval, fmt.Sprintf("%s start interval", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeySchedulerDispatcherSchedulerScheduleInterval, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerDispatcherSchedulerScheduleInterval, flagPrefix), DefaultSchedulerDispatcherSchedulerScheduleInterval, fmt.Sprintf("%s schedule interval", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeySchedulerDispatcherSchedulerStopTimeout, keyPrefix), fmt.Sprintf(TemplateFlagSchedulerDispatcherSchedulerStopTimeout, flagPrefix), DefaultSchedulerDispatcherStopTimeout, fmt.Sprintf("%s stop timeout", flagPrefix)),
	}
}
