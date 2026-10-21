package worker

import (
	"strings"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// Константы дефолтов для внутренней защиты рантайм-компонента (независимые от пакета config)
const (
	DefaultSchedulerStartInterval    time.Duration = time.Second
	DefaultSchedulerScheduleInterval time.Duration = time.Second * 5
	DefaultSchedulerStopTimeout      time.Duration = time.Second * 5
)

// PeriodicSchedulerOption определяет функциональный тип для конфигурации опций (Fluent API).
type PeriodicSchedulerOption func(*PeriodicSchedulerOptions)

// PeriodicSchedulerOptions содержит параметры рантайма, необходимые для безопасной сборки и работы scheduler.
type PeriodicSchedulerOptions struct {
	// Name Наименование воркера
	Name string

	// StartInterval Первичная задержка (холодное смещение) перед самым первым тиком таймера
	StartInterval time.Duration

	// ScheduleInterval Фиксированный интервал периодического повторения задач (период)
	ScheduleInterval time.Duration

	// StopTimeout Временной лимит (таймаут) на мягкое завершение активной итерации обработчика
	StopTimeout time.Duration

	// Logger логгер
	Logger logger.Logger

	// TimerDispatcher обработчик события таймер
	TimerDispatcher TimerDispatcher
}

// NewPeriodicSchedulerOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewPeriodicSchedulerOptions() *PeriodicSchedulerOptions {
	return &PeriodicSchedulerOptions{
		StartInterval:    DefaultSchedulerStartInterval,
		ScheduleInterval: DefaultSchedulerScheduleInterval,
		StopTimeout:      DefaultSchedulerStopTimeout,
	}
}

// Validate проверяет корректность абсолютно всех опций рантайма перед сборкой scheduler.
func (bso *PeriodicSchedulerOptions) Validate() error {
	if strings.TrimSpace(bso.Name) == "" {
		return errs.NewTlCommonError("Validate", "name is required", nil)
	}
	if bso.StartInterval <= 0 {
		return errs.NewTlCommonError("Validate", "start interval is required", nil)
	}
	if bso.ScheduleInterval <= 0 {
		return errs.NewTlCommonError("Validate", "schedule interval is required", nil)
	}
	if bso.StopTimeout <= 0 {
		return errs.NewTlCommonError("Validate", "stop timeout is required", nil)
	}
	if utils.IsNil(bso.TimerDispatcher) {
		return errs.NewTlCommonError("Validate", "timer dispatcher is required", nil)
	}
	if utils.IsNil(bso.Logger) {
		return errs.NewTlCommonError("Validate", "logger is required", nil)
	}

	return nil
}

// ====================================================================
// Fluent API методы для сборки опций scheduler
// ====================================================================

// WithPeriodicSchedulerName настраивает наименование
func WithPeriodicSchedulerName(name string) PeriodicSchedulerOption {
	return func(options *PeriodicSchedulerOptions) {
		options.Name = name
	}
}

// WithPeriodicSchedulerStartInterval настраивает первичную задержку (холодное смещение) перед самым первым тиком таймера
func WithPeriodicSchedulerStartInterval(interval time.Duration) PeriodicSchedulerOption {
	return func(options *PeriodicSchedulerOptions) {
		options.StartInterval = interval
	}
}

// WithPeriodicSchedulerScheduleInterval настраивает фиксированный интервал периодического повторения задач (период)
func WithPeriodicSchedulerScheduleInterval(interval time.Duration) PeriodicSchedulerOption {
	return func(options *PeriodicSchedulerOptions) {
		options.ScheduleInterval = interval
	}
}

// WithPeriodicSchedulerStopTimeout настраивает временной лимит (таймаут) на мягкое завершение активной итерации обработчика
func WithPeriodicSchedulerStopTimeout(timeout time.Duration) PeriodicSchedulerOption {
	return func(options *PeriodicSchedulerOptions) {
		options.StopTimeout = timeout
	}
}

// WithPeriodicSchedulerLogger настраивает scheduler logger
func WithPeriodicSchedulerLogger(logger logger.Logger) PeriodicSchedulerOption {
	return func(options *PeriodicSchedulerOptions) {
		options.Logger = logger
	}
}

// WithPeriodicSchedulerTimerDispatcher настраивает обработчик события
func WithPeriodicSchedulerTimerDispatcher(timerDispatcher TimerDispatcher) PeriodicSchedulerOption {
	return func(options *PeriodicSchedulerOptions) {
		options.TimerDispatcher = timerDispatcher
	}
}
