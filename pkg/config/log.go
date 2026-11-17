package config

import (
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги подсистемы структурированного логирования.
const (
	FlagLogLevel  string = "log-level"  // CLI-флаг глубины логирования (debug/info/warn/error)
	FlagLogFormat string = "log-format" // CLI-флаг формата вывода (json/console)
)

// ====================================================================
// logging
// ====================================================================
// Строковые ключи Viper-маппинга для иерархического связывания параметров логирования.
const (
	KeyLogLevel  string = "log.level"
	KeyLogFormat string = "log.format"
)

// ====================================================================
// logger defaults
// ====================================================================
// Настройки по умолчанию для подсистемы логирования.
const (
	DefaultLogLevel  string = "info"    // Базовый уровень фильтрации логов фреймворка
	DefaultLogFormat string = "console" // Базовый формат вывода (console для локальной разработки)
)

// LogConfig — уровни и формат логирования
//
// Инкапсулирует конфигурационные параметры подсистемы логирования фреймворка,
// включая глубину фильтрации (Level), формат кодирования (Format) и пути вывода.
// ToDo: Need refactoring!
type LogConfig struct {
	Level    string `mapstructure:"level" json:"level,omitempty" yaml:"level,omitempty" toml:"level,omitempty"`     // debug, info, warn, error
	Format   string `mapstructure:"format" json:"format,omitempty" yaml:"format,omitempty" toml:"format,omitempty"` // json, console
	FilePath string `mapstructure:"file_path" json:"file_path,omitempty" yaml:"file_path,omitempty" toml:"file_path,omitempty"`
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
	if strings.TrimSpace(lc.Level) == "" {
		return errs.NewConfigValidateError("log", "level", "must not be empty", nil)
	}
	if strings.TrimSpace(lc.Format) == "" {
		return errs.NewConfigValidateError("log", "format", "must not be empty", nil)
	}

	return nil
}

// BuildLogSettings log settings builder
func BuildLogSettings() Settings {
	return Settings{
		NewSetting(KeyLogLevel, FlagLogLevel, DefaultLogLevel, "log level"),
		NewSetting(KeyLogFormat, FlagLogFormat, DefaultLogFormat, "log format"),
	}
}
