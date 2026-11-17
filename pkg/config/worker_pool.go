package config

import (
	"fmt"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// flags templates
const (
	// TemplateFlagPoolWorkerCount кол-во обработчиков
	TemplateFlagPoolWorkerCount string = "%s-worker-count"

	// TemplateFlagPoolDataCapacity емкость внутреннего фонового буфера
	TemplateFlagPoolDataCapacity string = "%s-data-capacity"

	// TemplateFlagPoolCompleteProcess признак завершения обработки при закрытии
	TemplateFlagPoolCompleteProcess string = "%s-complete-process"

	// TemplateFlagPoolStopTimeout время на безопасную остановку и корректное завершение
	TemplateFlagPoolStopTimeout string = "%s-stop-timeout"
)

// keys templates
const (
	// TemplateKeyPoolWorkerCount кол-во обработчиков
	TemplateKeyPoolWorkerCount string = "%s.worker-count"

	// TemplateKeyPoolDataCapacity емкость внутреннего фонового буфера
	TemplateKeyPoolDataCapacity string = "%s.data_capacity"

	// TemplateKeyPoolCompleteProcess признак завершения обработки при закрытии
	TemplateKeyPoolCompleteProcess string = "%s.complete_process"

	// TemplateKeyPoolStopTimeout время на безопасную остановку и корректное завершение
	TemplateKeyPoolStopTimeout string = "%s.stop_timeout"
)

// ====================================================================
// Worker Pool Defaults
// ====================================================================
// Системные параметры по умолчанию для worker pool
const (
	// DefaultPoolWorkerCount - кол-во обработчиков
	DefaultPoolWorkerCount int = 2

	// DefaultPoolDataCapacity - емкость внутреннего фонового буфера
	DefaultPoolDataCapacity int = 100

	// DefaultPoolCompleteProcess - признак завершения обработки при закрытии
	DefaultPoolCompleteProcess bool = true

	// DefaultPoolStopTimeout - время на безопасную остановку и корректное завершение
	DefaultPoolStopTimeout time.Duration = 5 * time.Second
)

// WorkerPoolConfig инкапсулирует конфигурационные параметры емкости и политик останова пула воркеров.
type WorkerPoolConfig struct {
	// WorkerCount Количество параллельно запущенных горутин-обработчиков
	WorkerCount int `mapstructure:"worker_count" json:"worker_count,omitempty" yaml:"worker_count,omitempty" toml:"worker_count,omitempty"`

	// DataCapacity Буферная емкость внутреннего канала задач (Backpressure window)
	DataCapacity int `mapstructure:"data_capacity" json:"data_capacity,omitempty" yaml:"data_capacity,omitempty" toml:"data_capacity,omitempty"`

	// CompleteProcess Флаг: вычитывать ли буфер до конца при закрытии канала (true) или тушить экстренно (false)
	CompleteProcess bool `mapstructure:"complete_process" json:"complete_process,omitempty" yaml:"complete_process,omitempty" toml:"complete_process,omitempty"`

	// StopTimeout Временной лимит (таймаут) на мягкое завершение обработки перед принудительным выходом
	StopTimeout time.Duration `mapstructure:"stop_timeout" json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty" toml:"stop_timeout,omitempty"`
}

// NewWorkerPoolConfig — полный конструктор структуры конфигурации worker pool.
func NewWorkerPoolConfig(
	workerCount int,
	dataCapacity int,
	completeProcess bool,
	stopTimeout time.Duration,
) *WorkerPoolConfig {
	return &WorkerPoolConfig{
		WorkerCount:     workerCount,
		DataCapacity:    dataCapacity,
		CompleteProcess: completeProcess,
		StopTimeout:     stopTimeout,
	}
}

// NewDefaultWorkerPoolConfig собирает конфигурацию worker pool по умолчанию с наполнением системными дефолтами.
func NewDefaultWorkerPoolConfig() *WorkerPoolConfig {
	return NewWorkerPoolConfig(
		DefaultPoolWorkerCount,
		DefaultPoolDataCapacity,
		DefaultPoolCompleteProcess,
		DefaultPoolStopTimeout,
	)
}

// Validate выполняет строгую семантическую и математическую валидацию параметров worker pool на этапе запуска (Bootstrap Phase).
func (wpc *WorkerPoolConfig) Validate() error {
	if wpc.WorkerCount <= 0 {
		return errs.NewConfigValidateError("worker pool", "worker count", "must be greater than 0", nil)
	}
	if wpc.DataCapacity <= 0 {
		return errs.NewConfigValidateError("worker pool", "data capacity", "must be greater than 0", nil)
	}
	if wpc.StopTimeout <= 0 {
		return errs.NewConfigValidateError("worker pool", "stop timeout", "must be greater than 0", nil)
	}

	return nil
}

// BuildPoolSettings pool settings builder
func BuildPoolSettings(keyPrefix, flagPrefix string) Settings {
	return Settings{
		NewSetting(fmt.Sprintf(TemplateKeyPoolWorkerCount, keyPrefix), fmt.Sprintf(TemplateFlagPoolWorkerCount, flagPrefix), DefaultPoolWorkerCount, fmt.Sprintf("%s worker count", flagPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyPoolDataCapacity, keyPrefix), fmt.Sprintf(TemplateFlagPoolDataCapacity, keyPrefix), DefaultPoolDataCapacity, fmt.Sprintf("%s data capacity", keyPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyPoolCompleteProcess, keyPrefix), fmt.Sprintf(TemplateFlagPoolCompleteProcess, keyPrefix), DefaultPoolCompleteProcess, fmt.Sprintf("%s complete process", keyPrefix)),
		NewSetting(fmt.Sprintf(TemplateKeyPoolStopTimeout, keyPrefix), fmt.Sprintf(TemplateFlagPoolStopTimeout, keyPrefix), DefaultPoolStopTimeout, fmt.Sprintf("%s stop timeout", keyPrefix)),
	}
}
