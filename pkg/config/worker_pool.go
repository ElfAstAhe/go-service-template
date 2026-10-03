package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// WorkerPoolConfig инкапсулирует конфигурационные параметры емкости и политик останова пула воркеров.
type WorkerPoolConfig struct {
	// WorkerCount Количество параллельно запущенных горутин-обработчиков
	WorkerCount int `mapstructure:"worker_count" json:"worker_count,omitempty" yaml:"worker_count,omitempty"`

	// DataCapacity Буферная емкость внутреннего канала задач (Backpressure window)
	DataCapacity int `mapstructure:"data_capacity" json:"data_capacity,omitempty" yaml:"data_capacity,omitempty"`

	// CompleteProcess Флаг: вычитывать ли буфер до конца при закрытии канала (true) или тушить экстренно (false)
	CompleteProcess bool `mapstructure:"complete_process" json:"complete_process,omitempty" yaml:"complete_process,omitempty"`

	// StopTimeout Временной лимит (таймаут) на мягкое завершение обработки перед принудительным выходом
	StopTimeout time.Duration `mapstructure:"stop_timeout" json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty"`
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
