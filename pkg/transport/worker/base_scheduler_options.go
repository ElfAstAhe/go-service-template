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
	defaultSchedulerStartInterval    time.Duration = time.Second
	defaultSchedulerScheduleInterval time.Duration = time.Second * 5
	defaultSchedulerStopTimeout      time.Duration = time.Second * 5
)

// BaseSchedulerOption определяет функциональный тип для конфигурации опций (Fluent API).
type BaseSchedulerOption func(*BaseSchedulerOptions)

// BaseSchedulerOptions содержит параметры рантайма, необходимые для безопасной сборки и работы scheduler.
type BaseSchedulerOptions struct {
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

// NewBaseSchedulerOptions создает структуру опций, сразу наполненную безопасными рантайм-дефолтами.
func NewBaseSchedulerOptions() *BaseSchedulerOptions {
	return &BaseSchedulerOptions{
		StartInterval:    defaultSchedulerStartInterval,
		ScheduleInterval: defaultSchedulerScheduleInterval,
		StopTimeout:      defaultSchedulerStopTimeout,
	}
}

// Validate проверяет корректность абсолютно всех опций рантайма перед сборкой scheduler.
func (bso *BaseSchedulerOptions) Validate() error {
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

// WithSchedulerName настраивает наименование
func WithSchedulerName(name string) BaseSchedulerOption {
	return func(options *BaseSchedulerOptions) {
		options.Name = name
	}
}

// WithSchedulerStartInterval настраивает первичную задержку (холодное смещение) перед самым первым тиком таймера
func WithSchedulerStartInterval(interval time.Duration) BaseSchedulerOption {
	return func(options *BaseSchedulerOptions) {
		options.StartInterval = interval
	}
}

// WithSchedulerScheduleInterval настраивает фиксированный интервал периодического повторения задач (период)
func WithSchedulerScheduleInterval(interval time.Duration) BaseSchedulerOption {
	return func(options *BaseSchedulerOptions) {
		options.ScheduleInterval = interval
	}
}

// WithSchedulerStopTimeout настраивает временной лимит (таймаут) на мягкое завершение активной итерации обработчика
func WithSchedulerStopTimeout(timeout time.Duration) BaseSchedulerOption {
	return func(options *BaseSchedulerOptions) {
		options.StopTimeout = timeout
	}
}

// WithSchedulerLogger настраивает scheduler logger
func WithSchedulerLogger(logger logger.Logger) BaseSchedulerOption {
	return func(options *BaseSchedulerOptions) {
		options.Logger = logger
	}
}

// WithSchedulerTimerDispatcher настраивает обработчик события
func WithSchedulerTimerDispatcher(timerDispatcher TimerDispatcher) BaseSchedulerOption {
	return func(options *BaseSchedulerOptions) {
		options.TimerDispatcher = timerDispatcher
	}
}
