package config

import (
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// LogConfig — уровни и формат логирования
//
// Инкапсулирует конфигурационные параметры подсистемы логирования фреймворка,
// включая глубину фильтрации (Level), формат кодирования (Format) и пути вывода.
// ToDo: Need refactoring!
type LogConfig struct {
	Level    string `mapstructure:"level" json:"level,omitempty" yaml:"level,omitempty"`    // debug, info, warn, error
	Format   string `mapstructure:"format" json:"format,omitempty" yaml:"format,omitempty"` // json, console
	FilePath string `mapstructure:"file_path" json:"file_path,omitempty" yaml:"file_path,omitempty,"`
}

// NewLogConfig — фабричный конструктор конфигурации подсистемы логирования.
func NewLogConfig(level, format string, filePath string) *LogConfig {
	return &LogConfig{
		Level:  level,
		Format: format,
	}
}

// NewDefaultLogConfig собирает конфигурацию логгера по умолчанию, наполняя её системными константными дефолтами.
func NewDefaultLogConfig() *LogConfig {
	return NewLogConfig(DefaultLogLevel, DefaultLogFormat, "")
}

// Validate осуществляет семантическую проверку конфигурации логирования на этапе запуска приложения (Bootstrap Phase).
// Предотвращает старт системы наблюдения с незаданными уровнями или пустыми форматами вывода.
func (lc *LogConfig) Validate() error {
	if lc.Level == "" {
		return errs.NewConfigValidateError("log", "level", "must not be empty", nil)
	}
	if lc.Format == "" {
		return errs.NewConfigValidateError("log", "format", "must not be empty", nil)
	}

	return nil
}
