package config

import (
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// CLI-флаги для настройки транспортных параметров и KeepAlive-политик gRPC-сервера.
const (
	FlagGRPCAddress          string = "grpc-address"            // CLI-флаг сетевого адреса gRPC слушателя
	FlagGRPCMaxConnIdle      string = "grpc-max-conn-idle"      // CLI-флаг лимита простоя HTTP/2 соединений
	FlagGRPCMaxConnAge       string = "grpc-max-conn-age"       // CLI-флаг лимита времени жизни коннекта (Ротация)
	FlagGRPCMaxConnAgeGrace  string = "grpc-max-conn-age-grace" // CLI-флаг Grace-периода мягкого закрытия старых gRPC запросов
	FlagGRPCTimeout          string = "grpc-timeout"            // CLI-флаг сетевого таймаута gRPC операций
	FlagGRPCKeepAliveTime    string = "grpc-keep-alive-time"    // CLI-флаг интервала пинга KeepAlive со стороны сервера
	FlagGRPCKeepAliveTimeout string = "grpc-keep-alive-timeout" // CLI-флаг таймаута ожидания ответа на KeepAlive пинг
	FlagGRPCShutdownTimeout  string = "grpc-shutdown-timeout"   // CLI-флаг таймаута мягкого гашения gRPC сервера
)

// ====================================================================
// gRPC
// ====================================================================
// Строковые ключи Viper-маппинга для иерархического связывания параметров gRPC сервера.
const (
	KeyGRPCAddress          string = "grpc.address"
	KeyGRPCMaxConnIdle      string = "grpc.max-conn-idle"
	KeyGRPCMaxConnAge       string = "grpc.max-conn-age"
	KeyGRPCMaxConnAgeGrace  string = "grpc-max-conn-age-grace"
	KeyGRPCTimeout          string = "grpc.timeout"
	KeyGRPCKeepAliveTime    string = "grpc-keep-alive-time"
	KeyGRPCKeepAliveTimeout string = "grpc-keep-alive-timeout"
	KeyGRPCShutdownTimeout  string = "grpc.shutdown_timeout"
)

// ====================================================================
// gRPC defaults
// ====================================================================
// Настройки по умолчанию и политики балансировки HTTP/2 соединений для gRPC-транспорта.
const (
	DefaultGRPCAddress string = "localhost:50051" // Сетевой интерфейс и порт по умолчанию для gRPC-слушателя

	// DefaultGRPCMaxConnIdle - Даем соединениям «отдохнуть», но не убиваем их сразу, снижая нагрузку на сокеты.
	DefaultGRPCMaxConnIdle time.Duration = 5 * time.Minute

	// DefaultGRPCMaxConnAge - Ротируем соединения раз в 20 минут для равномерной балансировки трафика внутри K8s.
	DefaultGRPCMaxConnAge time.Duration = 20 * time.Minute

	// DefaultGRPCMaxConnAgeGrace - Grace-период, чтобы активные gRPC-стримы успели довариться при ротации коннекта.
	DefaultGRPCMaxConnAgeGrace  time.Duration = 1 * time.Minute
	DefaultGRPCTimeout          time.Duration = 5 * time.Second     // Жесткий порог ожидания выполнения gRPC операции
	DefaultGRPCKeepAliveTime    time.Duration = 2 * 2 * time.Minute // Интервал превентивной отправки пингов для проверки живой TCP-сессии
	DefaultGRPCKeepAliveTimeout time.Duration = 20 * time.Second    // Время ожидания ответа на пинг до принудительного дропа сокета
	DefaultGRPCShutdownTimeout  time.Duration = 15 * time.Second    // Лимит времени на мягкую остановку gRPC обработчиков
)

// GRPCConfig — настройки gRPC сервера
//
// Инкапсулирует конфигурационные параметры gRPC-транспорта, включая адресацию,
// параметры управления жизненным циклом HTTP/2 соединений и KeepAlive-политики.
type GRPCConfig struct {
	Address         string        `mapstructure:"address" json:"address,omitempty" yaml:"address,omitempty" toml:"address,omitempty"`
	MaxConnIdle     time.Duration `mapstructure:"max_conn_idle" json:"max_conn_idle,omitempty" yaml:"max_conn_idle,omitempty" toml:"max_conn_idle,omitempty"`
	MaxConnAge      time.Duration `mapstructure:"max_conn_age" json:"max_conn_age,omitempty" yaml:"max_conn_age,omitempty" toml:"max_conn_age,omitempty"`
	MaxConnAgeGrace time.Duration `mapstructure:"max_conn_age_grace" json:"max_conn_age_grace,omitempty" yaml:"max_conn_age_grace,omitempty" toml:"max_conn_age_grace,omitempty"`
	Timeout         time.Duration `mapstructure:"timeout" json:"timeout,omitempty" yaml:"timeout,omitempty" toml:"timeout,omitempty"`
	// Настройки KeepAlive важны, чтобы соединения не "протухали" за балансировщиками
	KeepAliveTime    time.Duration `mapstructure:"keep_alive_time" json:"keep_alive_time,omitempty" yaml:"keep_alive_time,omitempty" toml:"keep_alive_time,omitempty"`
	KeepAliveTimeout time.Duration `mapstructure:"keep_alive_timeout" json:"keep_alive_timeout,omitempty" yaml:"keep_alive_timeout,omitempty" toml:"keep_alive_timeout,omitempty"`
	ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout,omitempty" yaml:"shutdown_timeout,omitempty" toml:"shutdown_timeout,omitempty"`
}

// NewGRPCConfig конструктор
func NewGRPCConfig(
	address string,
	maxConnIdle,
	maxConnAge,
	maxConnAgeGrace,
	timeout,
	keepAliveTime,
	keepAliveTimeout,
	shutdownTimeout time.Duration,
) *GRPCConfig {
	return &GRPCConfig{
		Address:          address,
		MaxConnIdle:      maxConnIdle,
		MaxConnAge:       maxConnAge,
		MaxConnAgeGrace:  maxConnAgeGrace,
		Timeout:          timeout,
		KeepAliveTime:    keepAliveTime,
		KeepAliveTimeout: keepAliveTimeout,
		ShutdownTimeout:  shutdownTimeout,
	}
}

// NewDefaultGRPCConfig констркутор настроек по умолчанию
func NewDefaultGRPCConfig() *GRPCConfig {
	return NewGRPCConfig(
		DefaultGRPCAddress,
		DefaultGRPCMaxConnIdle,
		DefaultGRPCMaxConnAge,
		DefaultGRPCMaxConnAgeGrace,
		DefaultGRPCTimeout,
		DefaultGRPCKeepAliveTime,
		DefaultGRPCKeepAliveTimeout,
		DefaultGRPCShutdownTimeout)
}

// Validate осуществляет семантическую проверку параметров gRPC-транспорта на этапе запуска приложения (Bootstrap Phase).
// Предотвращает запуск транспортного узла с пустым адресом слушателя или отрицательными таймаутами.
func (gc *GRPCConfig) Validate() error {
	if gc.Address == "" {
		return errs.NewConfigValidateError("gRPC", "address", "must not be empty", nil)
	}
	if gc.MaxConnIdle < 0 {
		return errs.NewConfigValidateError("gRPC", "max_conn_idle", "must be greater or equal zero", nil)
	}
	if gc.MaxConnAge < 0 {
		return errs.NewConfigValidateError("gRPC", "max_conn_age", "must be greater or equal zero", nil)
	}
	if gc.MaxConnAgeGrace < 0 {
		return errs.NewConfigValidateError("gRPC", "max_conn_age_grace", "must be greater or equal zero", nil)
	}
	if gc.Timeout < 0 {
		return errs.NewConfigValidateError("gRPC", "timeout", "must be greater or equal zero", nil)
	}
	if gc.KeepAliveTime < 0 {
		return errs.NewConfigValidateError("gRPC", "keepalive_time", "must be greater or equal zero", nil)
	}
	if gc.KeepAliveTimeout < 0 {
		return errs.NewConfigValidateError("gRPC", "keepalive_timeout", "must be greater or equal zero", nil)
	}
	if gc.ShutdownTimeout < 0 {
		return errs.NewConfigValidateError("gRPC", "shutdown_timeout", "must be greater or equal zero", nil)
	}

	return nil
}

// BuildDefaultGRPCSettings GRPC settings builder
func BuildDefaultGRPCSettings() Settings {
	return Settings{
		NewSetting(KeyGRPCAddress, FlagGRPCAddress, DefaultGRPCAddress, "gRPC address"),
		NewSetting(KeyGRPCMaxConnIdle, FlagGRPCMaxConnIdle, DefaultGRPCMaxConnIdle, "gRPC max connection idle time"),
		NewSetting(KeyGRPCMaxConnAge, FlagGRPCMaxConnAge, DefaultGRPCMaxConnAge, "gRPC max connection age timeout"),
		NewSetting(KeyGRPCMaxConnAgeGrace, FlagGRPCMaxConnAgeGrace, DefaultGRPCMaxConnAgeGrace, "gRPC max connection age grace timeout"),
		NewSetting(KeyGRPCTimeout, FlagGRPCTimeout, DefaultGRPCTimeout, "gRPC timeout"),
		NewSetting(KeyGRPCKeepAliveTime, FlagGRPCKeepAliveTime, DefaultGRPCKeepAliveTime, "gRPC keep alive time"),
		NewSetting(KeyGRPCKeepAliveTimeout, FlagGRPCKeepAliveTimeout, DefaultGRPCKeepAliveTimeout, "gRPC keep alive timeout"),
		NewSetting(KeyGRPCShutdownTimeout, FlagGRPCShutdownTimeout, DefaultGRPCShutdownTimeout, "gRPC shutdown timeout"),
	}
}
