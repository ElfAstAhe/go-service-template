package config

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// LogConfig — уровни и формат логирования
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, console
}

func NewLogConfig(level, format string) *LogConfig {
	return &LogConfig{
		Level:  level,
		Format: format,
	}
}

func NewDefaultLogConfig() *LogConfig {
	return NewLogConfig(defaultLogLevel, defaultLogFormat)
}

func (lc *LogConfig) Validate() error {
	if lc.Level == "" {
		return errs.NewConfigValidateError("log", "level", "must not be empty", nil)
	}
	if lc.Format == "" {
		return errs.NewConfigValidateError("log", "format", "must not be empty", nil)
	}

	return nil
}
